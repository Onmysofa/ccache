package ccache

func evalLFU(i *Item) int32 {
	return i.accCount
}
