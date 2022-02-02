asertx developer start --src-path ~/aserto-dev/aserto-go/examples/middleware/policy/ local
asertx directory load-users --provider json --file ~/aserto-dev/aserto-go/examples/middleware/users.json --authorizer
localhost:8282 --incl-user-ext
