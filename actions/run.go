package actions

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/loadimpact/speedboat/actions/registry"
	"github.com/loadimpact/speedboat/common"
	"github.com/loadimpact/speedboat/master"
	"github.com/loadimpact/speedboat/message"
	"github.com/loadimpact/speedboat/runner"
	"github.com/loadimpact/speedboat/runner/js"
	"github.com/loadimpact/speedboat/worker"
	"io/ioutil"
)

func init() {
	registry.RegisterCommand(cli.Command{
		Name:   "run",
		Usage:  "Runs a load test",
		Action: actionRun,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "script",
				Usage: "Script file to run",
			},
		},
	})
	registry.RegisterProcessor(func(*worker.Worker) master.Processor {
		return &RunProcessor{}
	})
}

type RunProcessor struct{}

func (p *RunProcessor) Process(msg message.Message) <-chan message.Message {
	ch := make(chan message.Message)

	go func() {
		defer close(ch)

		switch msg.Type {
		case "run.run":
			filename := msg.Fields["filename"].(string)
			src := msg.Fields["src"].(string)

			log.WithFields(log.Fields{
				"filename": filename,
				"src":      src,
			}).Debug("Source")

			var r runner.Runner = nil

			r, err := js.New()
			if err != nil {
				ch <- message.NewToClient("run.error", message.Fields{"error": err})
				ch <- message.NewToClient("run.end", message.Fields{})
				break
			}

			for res := range r.Run(filename, src) {
				switch res.Type {
				case "log":
					ch <- message.NewToClient("run.log", message.Fields{
						"time": res.LogEntry.Time,
						"text": res.LogEntry.Text,
					})
				}
			}
			ch <- message.NewToClient("run.end", message.Fields{})
		}
	}()

	return ch
}

func actionRun(c *cli.Context) {
	client, _ := common.MustGetClient(c)
	in, out, errors := client.Connector.Run()

	if !c.IsSet("script") {
		log.Fatal("No script file specified!")
	}

	filename := c.String("script")
	srcb, err := ioutil.ReadFile(filename)
	src := string(srcb)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"filename": filename,
		}).Fatal("Couldn't read script")
	}

	out <- message.NewToWorker("run.run", message.Fields{
		"filename": c.String("script"),
		"src":      src,
	})

readLoop:
	for {
		select {
		case msg := <-in:
			switch msg.Type {
			case "run.log":
				log.WithFields(log.Fields{
					"time": msg.Fields["time"],
					"text": msg.Fields["text"],
				}).Info("Test Log")
			case "run.end":
				log.Info("-- Test End --")
				break readLoop
			}
		case err := <-errors:
			log.WithError(err).Error("Ping failed")
		}
	}
}