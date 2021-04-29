package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, j := range jobs {

		wg.Add(1)

		out := make(chan interface{})
		go execute(j, in, out, wg)
		in = out
	}
	wg.Wait()
}

func execute(j job, in, out chan interface{}, wg *sync.WaitGroup)  {
	j(in, out)
	close(out)
	wg.Done()
}


func SingleHash(in chan interface{}, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mut := &sync.Mutex{}
	for input :=  range in {
		wg.Add(1)
		go func(input int) {
			data := strconv.Itoa(input)
			mut.Lock()
			datamd5 := DataSignerMd5(data)
			mut.Unlock()
			extraChan := make(chan string)
			go func(out chan string) {
				out <-  DataSignerCrc32(datamd5)
				close(out)
			}(extraChan)
			res := DataSignerCrc32(data)
			out <- res + "~" + <-extraChan
			wg.Done()
		}(input.(int))
	}
	wg.Wait()
}

func MultiHash(in chan interface{}, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input :=  range in {
		wg.Add(1)
		go multiHash(input.(string), wg, out)
	}
	wg.Wait()
}

func multiHash(input string, wg *sync.WaitGroup, out chan interface{}){
	defer wg.Done()
	arr := make([]string, 6)
	wg2 := &sync.WaitGroup{}

	for i := 0; i < 6; i++ {
		wg2.Add(1)
		data := strconv.Itoa(i) + input
		go func(index int,  data string) {
			defer wg2.Done()
			arr[index] = DataSignerCrc32(data)
		}(i, data)
	}
	wg2.Wait()
	result := strings.Join(arr, "")
	out <- result
}

func CombineResults(in chan interface{}, out chan interface{}) {
	var arr [] string
	for input := range in{
		arr = append(arr, input.(string))
	}
	sort.Strings(arr)
	res := strings.Join(arr, "_")
	out <- res
}


