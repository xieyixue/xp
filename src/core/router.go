package core

type Router struct {
	Router map[string]map[string]int
}

func (r Router) Set() {
	r.Router = make(map[string]map[string]int)
}

func (r Router) Get(addr string) (target int) {
	target = r.Router[addr]["target"]
	return
}

func (r Router) Add(addr string, target int){
	m := make(map[string]int)
	m["target"] = target
	r.Router[addr] = m
}