package cli

import (
	"fmt"
	"sort"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	blockListCommand = blockCommands.Command("list", "List objects").Alias("ls")
	blockListKind    = blockListCommand.Flag("kind", "Block kind").Default("all").Enum("all", "physical", "packed", "nonpacked", "packs")
	blockListLong    = blockListCommand.Flag("long", "Long output").Short('l').Bool()
	blockListPrefix  = blockListCommand.Flag("prefix", "Prefix").String()
	blockListSort    = blockListCommand.Flag("sort", "Sort order").Default("name").Enum("name", "size", "time", "none")
	blockListReverse = blockListCommand.Flag("reverse", "Reverse sort").Short('r').Bool()
)

func runListBlocksAction(context *kingpin.ParseContext) error {
	rep := mustOpenRepository(nil)
	defer rep.Close()

	blocks := rep.Blocks.ListBlocks(*blockListPrefix, *blockListKind)
	maybeReverse := func(b bool) bool { return b }

	if *blockListReverse {
		maybeReverse = func(b bool) bool { return !b }
	}

	switch *blockListSort {
	case "name":
		sort.Slice(blocks, func(i, j int) bool { return maybeReverse(blocks[i].BlockID < blocks[j].BlockID) })
	case "size":
		sort.Slice(blocks, func(i, j int) bool { return maybeReverse(blocks[i].Length < blocks[j].Length) })
	case "time":
		sort.Slice(blocks, func(i, j int) bool { return maybeReverse(blocks[i].TimeStamp.Before(blocks[j].TimeStamp)) })
	}

	for _, b := range blocks {
		if b.Error != nil {
			return b.Error
		}

		if *blockListLong {
			fmt.Printf("%-34v %10v %v\n", b.BlockID, b.Length, b.TimeStamp.Local().Format(timeFormat))
		} else {
			fmt.Printf("%v\n", b.BlockID)
		}
	}

	return nil
}

func init() {
	blockListCommand.Action(runListBlocksAction)
}