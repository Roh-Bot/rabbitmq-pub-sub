package functions

const (
	GenerateAndGetQueueName = `SELECT * from rabbitmq.generate_and_add_queue()`
	GetAllQueues            = `SELECT * from rabbitmq.get_all_queues()`
	DeleteQueue             = `SELECT * from rabbitmq.delete_queue($1)`
	TruncateQueues          = `SELECT truncate_queues_table()`
)
