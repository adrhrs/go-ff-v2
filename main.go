package main

import (
	"fmt"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func main() {
	initFlag()

}

func initFlag() {
	ffclient.Init(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever: &fileretriever.Retriever{
			Path: "flag.yaml",
		},
	})

	user := ffcontext.NewEvaluationContext("1")
	allFlagsState := ffclient.AllFlagsState(user)
	fmt.Println(allFlagsState.GetFlags()["aa"].Value)

	user1 := ffcontext.
		NewEvaluationContextBuilder("4").
		AddCustom("cohort", "xy").
		Build()
	allFlagsState = ffclient.AllFlagsState(user1)
	fmt.Println(allFlagsState.GetFlags()["aa"].Value)

	defer ffclient.Close()
}
