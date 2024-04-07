package bimap

type M[K, V comparable] struct {
	kv map[K]V
	vk map[V]K
}

func New[K, V comparable](values map[K]V) *M[K, V] {
	kv := make(map[K]V, len(values))
	vk := make(map[V]K, len(values))
	for k, v := range values {
		kv[k] = v
		vk[v] = k
	}

	return &M[K, V]{
		kv: kv,
		vk: vk,
	}
}

func (bm *M[K, V]) GetByKey(k K) (V, bool) {
	out, ok := bm.kv[k]
	return out, ok
}

func (bm *M[K, V]) GetByValue(v V) (K, bool) {
	out, ok := bm.vk[v]
	return out, ok
}

func (bm *M[K, V]) Put(k K, v V) {
	bm.kv[k] = v
	bm.vk[v] = k
}
