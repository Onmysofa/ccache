package ccache

import "time"

func evalLFU(i *Item) float64 {
	return float64(i.accCount)
}

func evalLRU(i *Item) float64 {
	return float64(i.accessTs.Unix())
}

func evalHyperbolic(i *Item) float64 {
	t := time.Now().Sub(i.createTS)
	return float64(i.accCount) / float64(t) / float64(i.size)
}

func evalOursH1(i *Item) float64 {
	t := time.Now().Sub(i.createTS)
	return float64(i.accCount) / float64(t) / i.reqInfo.ReqSize
}

func evalOursH2(i *Item) float64 {
	t := time.Now().Sub(i.createTS)
	return float64(i.accCount) / float64(t) / i.reqInfo.MissingSize
}