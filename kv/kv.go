package kv

const (
	INMEMORY Kind = "in-memory"
)

var supportedKVs map[string]Kind = map[string]Kind{
	"in-memory": INMEMORY,
}

type KV interface {
	Put(k string, v *Value) error
	Get(k string) (*Value, error)
	Delete(k string) error
}

type Kind string

func KVFromName(name string) (Kind, error) {
	kind, ok := supportedKVs[name]
	if !ok {
		return kind, NewNotImplementedError(name)
	}
	return kind, nil
}

func NewKV(kind Kind) (KV, error) {
	var kv KV
	switch kind {
	case INMEMORY:
		kv = &InMemory{Db: make(map[string]*Value)}
		break
	}
	return kv, nil
}
