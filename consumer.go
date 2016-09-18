package main

type Consumer struct {
	Progress chan int64
}

func NewConsumer() (c *Consumer) {
	c = new(Consumer)
	c.Progress = make(chan int64)
	return c
}
