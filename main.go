package main

import (
	"github.com/sirupsen/logrus"

	"github.com/hiltpold/lakelandcup-fantasy-service/commands"
)

func main() {
	if err := commands.RootCommand().Execute(); err != nil {
		logrus.Fatal(err)
	}
}
