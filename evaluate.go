package ccache

func evalLFU(i *Item) int64 {
	return i.accCount
}

func evalLRU(i *Item) int64 {
	return int64(i.accessTs.Unix())
}

