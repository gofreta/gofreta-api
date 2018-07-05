var nowTimestamp = Date.now() / 1000 << 0;

// insert user "admin" with password "123456"
if (!db.user.find().length()) {
    var adminUser = {"username": "admin", "email": "admin@example.com", "status": "active", "password_hash": "$2a$12$rdX7N6gpAzKJ/7DzCMyVdeRaTUv6faL6GxhTODzlJcuDHRf4hedoO", "reset_password_hash": "", "access": {"user": ["index", "view", "create", "update", "delete"], "key": ["index", "view", "create", "update", "delete"], "language": ["create", "update", "delete"], "media": ["index", "view", "upload", "update", "delete", "replace"], "collection": ["index", "view", "create", "update", "delete"]}, "created": nowTimestamp, "modified": nowTimestamp};
    db.user.insert(adminUser);
}

// insert English(en) language
if (!db.language.find().length()) {
    var language = {"title": "English", "locale": "en", "created": nowTimestamp, "modified": nowTimestamp};
    db.language.insert(language);
}
