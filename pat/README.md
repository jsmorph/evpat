# Event pattern (matching)

A naive implementation of [Amazon EventBridge event
   patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html).

The tests include an option to use the AWS SDK to [test
matching](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/eventbridge#Client.TestEventPattern).

## ToDo

- [ ] Null
- [ ] CIDR
- [ ] Number comparison is supposed to be by string comparison?

    > For numbers, EventBridge uses string representation. For
    > example, 300, 300.0, and 3.0e2 are not considered equal.
	
- [ ] Many more test cases
- [x] `anything-but`
- [x] Test cases as JSON
- [ ] Benchmarks
- [x] Array values
- [ ] Better documentation since IMHO the AWS [docs](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html) aren't great
