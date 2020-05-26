package main

import (
	"github.com/cheekybits/genny/generic"
    "github.com/dghubble/go-twitter/twitter"
	"sync"
)

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "DataType=twitter.Tweet,twitter.User,AgriusUserStatus"

type DataType generic.Type

func mergeDataTypeChans(cs ...<-chan *DataType) <-chan *DataType {
    var wg sync.WaitGroup
    out := make(chan *DataType)

    output := func(c <-chan *DataType) {
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


type DataTypeWork interface {
	Do(client *twitter.Client, inchan chan DataTypeWork, outchan chan *DataType, errs chan error)
}

func (ac *AgriusClient) DataTypeWorker(inchan chan DataTypeWork, errs chan error) chan *DataType {
	outchan := make(chan *DataType)

	go func() {
		defer close(outchan)

		for w := range inchan {
			w.Do(ac.client, inchan, outchan, errs)
		}
	}()

	return outchan
}


func DataTypeWorkerPool(cnf Config, userTokens map[string]string, errs chan error) (chan<- DataTypeWork, <-chan *DataType) {

	rech := make(chan DataTypeWork)
	outs := []<-chan *DataType{}

	for k, s := range userTokens {
		ac := newAgriusClient(cnf.TwitterToken, cnf.TwitterSecret, k, s)
		out := ac.DataTypeWorker(rech, errs)
		outs = append(outs, out)
	}

	outch := mergeDataTypeChans(outs...)

	return rech, outch
}


func DataTypeMaster(cnf Config, userTokens map[string]string, inch <-chan DataTypeWork, errs chan error) <-chan *DataType {
	outs := make(chan *DataType)

	go func() {
		defer close(outs)
		for w := range inch {

			rech, out := DataTypeWorkerPool(cnf, userTokens, errs)

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


func DataTypeMasterPool(cnf Config, N int) (chan<- DataTypeWork, <-chan *DataType, chan error) {
	userTokens := getTokens(cnf.TokenBeastLocation, cnf.TokenBeastSecret)

	inch := make(chan DataTypeWork)
	errs := make(chan error)
	outs := []<-chan *DataType{}

	for i := 0; i < N; i++ {
		out := DataTypeMaster(cnf, userTokens, inch, errs)
		outs = append(outs, out)
	}

	outch := mergeDataTypeChans(outs...)	

	return inch, outch, errs
}

