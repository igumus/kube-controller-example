package controller

type event struct {
	added bool
	obj   interface{}
}

func NewEvent(added bool, obj interface{}) event {
	return event{
		added: added,
		obj:   obj,
	}
}
