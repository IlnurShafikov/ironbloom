package pairs

type Pair struct {
	Key   int
	Value int
}

type Pairs []Pair

func (kv Pairs) GetValueByKey(key int) (int, bool) {
	for i := range kv {
		curr := kv[i]

		if curr.Key == key {
			return curr.Value, true
		}
	}

	return 0, false
}

func (kv Pairs) GetIDByValue(value int) (int, bool) {
	for i := range kv {
		curr := kv[i]

		if curr.Value == value {
			return curr.Key, true
		}
	}

	return 0, false
}

func (kv Pairs) GetByIndex(index int) (Pair, bool) {
	if index < 0 || index >= len(kv) {
		return Pair{Key: 0, Value: 0}, false
	}

	return kv[index], true
}
