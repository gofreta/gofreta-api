Gofreta REST API server
======================================================================

> **The project is still in very early development stage and it is not recommended to use in production environment, unless you really know what you are doing!**

"Go + Mongo" REST API server implementation for [Gofreta](https://gofreta.com).

- [Install](#install)
- [Development](#development)
- [Configurations](#configurations)
- [API Reference](#api-reference)

## Install

- Install [MongoDB 3.2+](https://www.mongodb.com/download-center?jmp=nav#community) and after that apply the following command to insert an initial user and language items (you can change them later):
  ```bash
  mongo localhost/gofreta --eval '
  var nowTimestamp = Date.now() / 1000 << 0;
  // insert user "admin" with password "123456"
  var adminUser = {"username": "admin", "email": "admin@example.com", "status": "active", "password_hash": "$2a$12$rdX7N6gpAzKJ/7DzCMyVdeRaTUv6faL6GxhTODzlJcuDHRf4hedoO", "reset_password_hash": "", "access": {"user": ["index", "view", "create", "update", "delete"], "key": ["index", "view", "create", "update", "delete"], "language": ["create", "update", "delete"], "media": ["index", "view", "upload", "update", "delete", "replace"], "collection": ["index", "view", "create", "update", "delete"]}, "created": nowTimestamp, "modified": nowTimestamp};
  db.user.insert(adminUser);
  // insert English(en) language
  var language = {"title": "English", "locale": "en", "created": nowTimestamp, "modified": nowTimestamp};
  db.language.insert(language);
  '
  ```

- Download the latest [Gofreta binary release](https://github.com/gofreta/gofreta-api/releases) and place it on your server.
  Execute the binary and specify the environment configuration file (see [Configurations](#configurations)):
  ```bash
  ./gofreta -config="/path/to/config.yaml"
  ```

That's it :). Check the API Reference documentation for info how to use the API.


## Development

Requirements:
- [Go 1.6+](https://golang.org/doc/install)
- [MongoDB 3.2+](https://www.mongodb.com/download-center?jmp=nav#community)

To download and install the package, execute the following command:
```bash
# install the application
go get github.com/gofreta/gofreta-api

# install glide (a vendoring and dependency management tool), if you don't have it yet
go get -u github.com/Masterminds/glide

# fetch the dependent packages
cd $GOPATH/gofreta/gofreta-api
$GOPATH/bin/glide install
```

Now you can build and run the application by running the following command:
```bash
# navigate to the applicatin directory
cd $GOPATH/gofreta/gofreta-api

# run the application with the default configurations
go run server.go

# run the application with user specified configuration file
go run server.go -config="/path/to/config.yaml"
```


## Configurations

Here is a yaml config file with the default application configurations:
You can <strong>extend it</strong> by using the `-config` flag (eg. `-config="/path/to/config.yaml"`).

```yaml
# the API base http server address
host: "http://localhost:8090"

# the Data Source Name for the database
dsn: "localhost/gofreta"

# mail server settings (if `host` is empty no emails will be send)
mailer:
  host:     ""
  username: ""
  password: ""
  port:     25

# these are secret keys used for JWT signing and verification
jwt:
  verificationKey: "__your_key__"
  signingKey:      "__your_key__"
  signingMethod:   "HS256"

# user auth token session duration (in hours)
userTokenExpire: 72

# reset password settings
resetPassword:
  # user reset password token secret
  secret: "__your_secret__"
  # user reset password token valid duration time (in hours)
  expire: 2
  # if not empty, the link will be included in the reset password email
  # (use `<hash>` as a placeholder for the reset password token, eg. `http://example.com/reset-password/<hash>`)
  pageLink: ""

# pagination settings
pagination:
  defaultLimit: 15
  maxLimit:     100

# upload settings
upload:
  maxSize: 5
  thumbs:  ["100x100", "300x300"]
  dir:     "./uploads"
  url:     "http://localhost:8090/upload"

# system email addresses
emails:
  noreply: "noreply@example.com"
  support: "support@example.com"
```

> I recommend you to double check the following parameters: `host`, `dsn`, `mailer`, `jwt`, `resetPassword.secret` and `upload`.


## API Reference

Detailed info and response examples are available at the offical Gofreta docs - https://gofreta.com/docs


## Credits

Gofreta REST API is part from [Gofreta](https://gofreta.com) - an Open Source project licensed under the [BSD3-License](LICENSE.md).

Help us improve and continue the project development - [https://gofreta.com/support-us](https://gofreta.com/support-us)
