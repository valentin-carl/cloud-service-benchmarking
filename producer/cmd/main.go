package main

import (
	"benchmark/lib"
	"benchmark/lib/utils"
	"log"
	"producer/pkg/workload"
)

func main() {

	lib.SayHi()

	//	messageSize := uint(math.Pow(2, 10))
	messageSize := uint(32)
	log.Println(messageSize)

	/*conf := config.Load("./config.json")

	g := workload.NewGenerator(
		messageSize,
		40,
		conf,
	)
	log.Println(g)

	msgs := g.GenerateMessages()
	for _, msg := range msgs {
		log.Println(string(msg))
	}

	// todo check if workload-something in config is empty => generate new workload
	name, err := g.GetWorkloadName()
	utils.Handle(err)
	utils.Handle(g.Store(msgs, name, 10))

	// increment experiment id for consistency with other nodes + filenames of stored workload
	err = conf.IncrementExperimentId("./config.json")
	utils.Handle(err)*/

	wls, err := workload.LoadWorkloads(messageSize, "workload-run-3")
	utils.Handle(err)

	log.Println("aaa", len(wls))
	for i, wl := range wls {
		log.Println(i, "wl ========")
		for _, msg := range wl {
			log.Println(string(msg))
		}
	}
}

/*
todo
- in config einstellen ob workload geladen oder neu generiert werden soll
	=> producer soll immer als in put workload/[][][]byte bekommen, dann auch worker aufteilen
- Wie sicherstellen dass es nur eine config.json gibt?
	=> for now: in root dir, für deployment env var mit config path?
	- weiteres problem, die dateistruktur wegen lib muss dann auch auf den nodes sein
		=> würde aber config.json auf root level einfach machen
- lib auch in consumer benutzen
*/
