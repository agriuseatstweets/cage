// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package main

import (
	"sync"

	"github.com/dghubble/go-twitter/twitter"
)

func mergeTwitterTweetChans(cs ...<-chan *twitter.Tweet) <-chan *twitter.Tweet {
	var wg sync.WaitGroup
	out := make(chan *twitter.Tweet)

	output := func(c <-chan *twitter.Tweet) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

type TwitterTweetWork interface {
	Do(client *twitter.Client, inchan chan TwitterTweetWork, outchan chan *twitter.Tweet, errs chan error)
}

func (ac *AgriusClient) TwitterTweetWorker(inchan chan TwitterTweetWork, errs chan error) chan *twitter.Tweet {
	outchan := make(chan *twitter.Tweet)

	go func() {
		defer close(outchan)

		for w := range inchan {
			w.Do(ac.client, inchan, outchan, errs)
		}
	}()

	return outchan
}

func TwitterTweetWorkerPool(cnf Config, userTokens map[string]string, errs chan error) (chan<- TwitterTweetWork, <-chan *twitter.Tweet) {

	rech := make(chan TwitterTweetWork)
	outs := []<-chan *twitter.Tweet{}

	for k, s := range userTokens {
		ac := newAgriusClient(cnf.TwitterToken, cnf.TwitterSecret, k, s)
		out := ac.TwitterTweetWorker(rech, errs)
		outs = append(outs, out)
	}

	outch := mergeTwitterTweetChans(outs...)

	return rech, outch
}

func TwitterTweetMaster(cnf Config, userTokens map[string]string, inch <-chan TwitterTweetWork, errs chan error) <-chan *twitter.Tweet {
	outs := make(chan *twitter.Tweet)

	go func() {
		defer close(outs)
		for w := range inch {

			rech, out := TwitterTweetWorkerPool(cnf, userTokens, errs)

			go func() {
				// kick off
				rech <- w
			}()

			// block until worker is finished
			for o := range out {
				outs <- o
			}
		}
	}()

	return outs
}

func TwitterTweetMasterPool(cnf Config, N int) (chan<- TwitterTweetWork, <-chan *twitter.Tweet, chan error) {
	userTokens := getTokens(cnf.TokenBeastLocation, cnf.TokenBeastSecret)

	inch := make(chan TwitterTweetWork)
	errs := make(chan error)
	outs := []<-chan *twitter.Tweet{}

	for i := 0; i < N; i++ {
		out := TwitterTweetMaster(cnf, userTokens, inch, errs)
		outs = append(outs, out)
	}

	outch := mergeTwitterTweetChans(outs...)

	return inch, outch, errs
}

func mergeTwitterUserChans(cs ...<-chan *twitter.User) <-chan *twitter.User {
	var wg sync.WaitGroup
	out := make(chan *twitter.User)

	output := func(c <-chan *twitter.User) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

type TwitterUserWork interface {
	Do(client *twitter.Client, inchan chan TwitterUserWork, outchan chan *twitter.User, errs chan error)
}

func (ac *AgriusClient) TwitterUserWorker(inchan chan TwitterUserWork, errs chan error) chan *twitter.User {
	outchan := make(chan *twitter.User)

	go func() {
		defer close(outchan)

		for w := range inchan {
			w.Do(ac.client, inchan, outchan, errs)
		}
	}()

	return outchan
}

func TwitterUserWorkerPool(cnf Config, userTokens map[string]string, errs chan error) (chan<- TwitterUserWork, <-chan *twitter.User) {

	rech := make(chan TwitterUserWork)
	outs := []<-chan *twitter.User{}

	for k, s := range userTokens {
		ac := newAgriusClient(cnf.TwitterToken, cnf.TwitterSecret, k, s)
		out := ac.TwitterUserWorker(rech, errs)
		outs = append(outs, out)
	}

	outch := mergeTwitterUserChans(outs...)

	return rech, outch
}

func TwitterUserMaster(cnf Config, userTokens map[string]string, inch <-chan TwitterUserWork, errs chan error) <-chan *twitter.User {
	outs := make(chan *twitter.User)

	go func() {
		defer close(outs)
		for w := range inch {

			rech, out := TwitterUserWorkerPool(cnf, userTokens, errs)

			go func() {
				// kick off
				rech <- w
			}()

			// block until worker is finished
			for o := range out {
				outs <- o
			}
		}
	}()

	return outs
}

func TwitterUserMasterPool(cnf Config, N int) (chan<- TwitterUserWork, <-chan *twitter.User, chan error) {
	userTokens := getTokens(cnf.TokenBeastLocation, cnf.TokenBeastSecret)

	inch := make(chan TwitterUserWork)
	errs := make(chan error)
	outs := []<-chan *twitter.User{}

	for i := 0; i < N; i++ {
		out := TwitterUserMaster(cnf, userTokens, inch, errs)
		outs = append(outs, out)
	}

	outch := mergeTwitterUserChans(outs...)

	return inch, outch, errs
}

func mergeAgriusUserStatusChans(cs ...<-chan *AgriusUserStatus) <-chan *AgriusUserStatus {
	var wg sync.WaitGroup
	out := make(chan *AgriusUserStatus)

	output := func(c <-chan *AgriusUserStatus) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

type AgriusUserStatusWork interface {
	Do(client *twitter.Client, inchan chan AgriusUserStatusWork, outchan chan *AgriusUserStatus, errs chan error)
}

func (ac *AgriusClient) AgriusUserStatusWorker(inchan chan AgriusUserStatusWork, errs chan error) chan *AgriusUserStatus {
	outchan := make(chan *AgriusUserStatus)

	go func() {
		defer close(outchan)

		for w := range inchan {
			w.Do(ac.client, inchan, outchan, errs)
		}
	}()

	return outchan
}

func AgriusUserStatusWorkerPool(cnf Config, userTokens map[string]string, errs chan error) (chan<- AgriusUserStatusWork, <-chan *AgriusUserStatus) {

	rech := make(chan AgriusUserStatusWork)
	outs := []<-chan *AgriusUserStatus{}

	for k, s := range userTokens {
		ac := newAgriusClient(cnf.TwitterToken, cnf.TwitterSecret, k, s)
		out := ac.AgriusUserStatusWorker(rech, errs)
		outs = append(outs, out)
	}

	outch := mergeAgriusUserStatusChans(outs...)

	return rech, outch
}

func AgriusUserStatusMaster(cnf Config, userTokens map[string]string, inch <-chan AgriusUserStatusWork, errs chan error) <-chan *AgriusUserStatus {
	outs := make(chan *AgriusUserStatus)

	go func() {
		defer close(outs)
		for w := range inch {

			rech, out := AgriusUserStatusWorkerPool(cnf, userTokens, errs)

			go func() {
				// kick off
				rech <- w
			}()

			// block until worker is finished
			for o := range out {
				outs <- o
			}
		}
	}()

	return outs
}

func AgriusUserStatusMasterPool(cnf Config, N int) (chan<- AgriusUserStatusWork, <-chan *AgriusUserStatus, chan error) {
	userTokens := getTokens(cnf.TokenBeastLocation, cnf.TokenBeastSecret)

	inch := make(chan AgriusUserStatusWork)
	errs := make(chan error)
	outs := []<-chan *AgriusUserStatus{}

	for i := 0; i < N; i++ {
		out := AgriusUserStatusMaster(cnf, userTokens, inch, errs)
		outs = append(outs, out)
	}

	outch := mergeAgriusUserStatusChans(outs...)

	return inch, outch, errs
}
