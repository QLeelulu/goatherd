package base

type Babylon struct {
    MyCapitain Captain
}

func (this *Babylon) Start() (err error) {
    var guards = this.MyCapitain.ListGuards()
    for _, guard := range guards {
        if err = guard.Listen(); err != nil {
            return
        }
    }
}
