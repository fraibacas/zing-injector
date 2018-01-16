# zing-injector
Tool to inject metrics or facts into a pubsub topic.
1) Build the executable `go build` if needed
2) Start pubsub emulator `gcloud beta emulators pubsub start`
3) Edit variables in file `env` and load them (`source env`) if needed
4) Send stuff from files in json:
    `./message_injector metrics metrics.json`
    `./message_injector facts facts.json`