package ccache

import "fmt"

func buildKey (backend, uri uint64) string {
	return fmt.Sprintf("%v:%v", backend, uri)
}

func parseKey(key string) (backend, uri uint64, err error) {
	if n,err := fmt.Sscanf(key, "%v:%v", &backend, &uri); n != 2 || err != nil {
		return 0, 0, err
	}

	return backend, uri, nil
}