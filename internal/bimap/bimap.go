package bimap

type Bimap[K, V comparable] struct {
	kv map[K]V
	vk map[V]K
}

func New[K, V comparable](values map[K]V) *Bimap[K, V] {
	vk := make(map[V]K, len(values))
	for k, v := range values {
		vk[v] = k
	}

	return &Bimap[K, V]{
		kv: values,
		vk: vk,
	}
}

func (bm *Bimap[K, V]) GetByKey(k K) (V, bool) {
	out, ok := bm.kv[k]
	return out, ok
}

func (bm *Bimap[K, V]) GetByValue(v V) (K, bool) {
	out, ok := bm.vk[v]
	return out, ok
}
