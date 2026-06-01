验证批量消费的正确方法：

  方法1：先堆积消息，再消费
  # 1. 停止消费者

  # 2. 快速发送大量消息（堆积到 Broker）
  go run tools/rmq_bench.go -count 10000 -rate 0