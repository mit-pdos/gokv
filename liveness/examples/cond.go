package examples

type CondVar struct {
	signal *AtomicUint64
}

func (c *CondVar) Wait() {
	for {
		if c.signal.Read() == 0 {
		}
	}
}
