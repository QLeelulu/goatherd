package collie

type Contex struct {
    *Collied
}

func NewContex() *Contex {
    return &Contex{
        Collied: NewCollied(),
    }
}
