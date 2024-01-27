package main

import (
	"fmt"
	"bufio"
	"os"
	"sync"
	"math"
	"math/rand"
)

var words []string
var length int

func readFile(filename string) []string {
	file, err := os.Open(filename)

	if err != nil {
		fmt.Println(err)
	}

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	var data []string

	for fileScanner.Scan() {
		new_word := fileScanner.Text()
		if len(new_word) == 5 {
			data = append(data, new_word)
		}
	}
	file.Close()
	return data
}

func isin(guess string, character string) int {
	for i := 0; i < 5; i++ {
		if string(guess[i]) == character {
			return i
		}
	}
	return -1
}


func word_score(guess string, one_word string) [5]int {
	taken := [5]bool{false, false, false, false, false}
	score := [5]int{-1,-1,-1,-1,-1}
	var i int
	for i = 0; i < 5; i++ {
		if string(guess[i]) == string(one_word[i]) {
			score[i] = 1
			taken[i] = true
		}
	}
	for i = 0; i < 5; i++ {
		index := isin(one_word, string(guess[i]))
		if index != -1 && !taken[index] {
			score[i] = 0
			taken[index] = true
		}
	}
	return score
}

func key_exists(scores map[[5]int]int, key [5]int) bool{
	for k := range scores {
		if k == key {
			return true
		}
	}
	return false
}

func entropy(scores map[[5]int]int) float64{
	var et float64 = 0
	for key := range scores {
		p := float64(scores[key]) / float64(len(words))
		et -= p * math.Log2(p)
	}
	return et
}

func score_all(guess string) float64 {
	scores_map := make(map[[5]int]int)
	for i := 0; i < len(words); i++ {
		if len(guess) == 5 && len(words[i]) == 5 {
			score := word_score(guess, words[i])
			if key_exists(scores_map, score) {
				scores_map[score]++
			} else {
				scores_map[score] = 1
			}
		}
	}
	return entropy(scores_map)
}

type result struct {
	word string
	result float64
}

func combination_generator(num_threads int) (<-chan string) {
	var i int
	ch := make(chan string, num_threads)
	go func() {
		for i = 0; i < len(words); i++ {
			ch <- words[i]
		}
		for i = 0; i < num_threads; i++ {
			ch <- "Done"
		}
		close(ch)
	}()
	return ch
}

func score_ch(in <-chan string, out chan <- result) {
	var msg string
	for {
		msg = <- in 
		if msg == "Done" {
			out <- result{word: "Done", result: 0}
			break
		} else {
			if len(msg) == 5 {
				out <- result{word: msg, result: score_all(msg)}
			}
		}
	}
}

func collect_values(in <-chan result, best_result *result, wg *sync.WaitGroup, num_threads int) {
	var msg result
	defer wg.Done()
	done_counter := 0
	for {
		msg = <- in
		if msg.word == "Done" {
			done_counter++
			if done_counter == num_threads {
				break
			}
		} else {
			if msg.result > (*best_result).result {
				*best_result = msg
			}
		}
	}
}

func is_match(word1 string, answers [5]int, word2 string) bool {
	taken := [5]bool{false, false, false, false, false}
	var valid bool = false
	var index int
	for i := 0; i < 5; i++ {
		index = isin(word2, string(word1[i]))
		if answers[i] == -1 && index != -1 {
			return false
		} else if answers[i] == 0 {
			if index == -1 {
				return false
			} else {
				for j := 0; j < 5; j++ {
					if word1[i] == word2[i] && !taken[j] {
						valid = true
					}
				}
				if !valid {
					return false
				}
			}
			taken[index] = true
		} else if answers[i] == 1 {
			if word1[i] != word2[i] {
				return false
			} else if !taken[index] {
				taken[index] = true
			}
		}
	}
	return true
}

func remove_words(word string, answers [5]int) {
	var new_words []string
	for i := 0; i < len(words); i++ {
		if is_match(word, answers, words[i]) && words[i] != word {
			new_words = append(new_words, words[i])
		}
	}
	words = new_words
}

func score_guess(guess string, word string) [5]int {
	taken := [5]bool{false, false, false, false, false}
	guess_result := [5]int{-1, -1, -1, -1, -1}
	if len(guess) != 5 {
		fmt.Println("Guess")
	}
	if len(word) != 5 {
		fmt.Println("Word")
	}
	for i := 0; i < 5; i++ {
		if guess[i] == word[i] {
			guess_result[i] = 1
			taken[i] = true
		} else {
			index := isin(word, string(guess[i]))
			if index != -1 {
				if !taken[index] {
					taken[index] = true
					guess_result[i] = 0
				}
			}
		}
	}
	return guess_result
}

func sum(scores_result [5]int) int {
	somme := 0
	for i := 0; i < 5; i++ {
		somme += scores_result[i]
	}
	return somme
}

func first_round(first_word string, word string) {
	guess_result := score_guess(first_word, word)
	fmt.Println(first_word)
	fmt.Println(guess_result)
	remove_words(best_result.word, guess_result)
}

func play_round(num_threads int, word string) bool {
	var wg sync.WaitGroup
	wg.Add(1)
	comb_ch := combination_generator(num_threads)
	score_channel := make(chan result, num_threads)
	for i := 0; i < num_threads; i++ {
		go score_ch(comb_ch, score_channel)
	}
	best_result := result{word : "", result : -1}
	go collect_values(score_channel, &best_result, &wg, num_threads)
	wg.Wait()
	close(score_channel)
	guess_result := score_guess(best_result.word, word)
	fmt.Println(best_result)
	fmt.Println(guess_result)
	remove_words(best_result.word, guess_result)
	return sum(guess_result) == 5
}

func main() {
	words = readFile("valid-wordle-words.txt")
	var word string
	for {
		word = words[rand.Intn(len(words))]
		if len(word) == 5 {
			break
		}
	}
	fmt.Println(word)
	first_round("tares", word)
	for i := 0; i < 5; i++ {
		if play_round(10, word) {
			break
		}
	}
}