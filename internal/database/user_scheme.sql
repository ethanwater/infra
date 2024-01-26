DROP TABLE IF EXISTS users;
CREATE TABLE users(
  id         INT UNIQUE NOT NULL,
  email      VARCHAR(128) UNIQUE,
  password   VARCHAR(128) NOT NULL,
  PRIMARY KEY (`id`)
);

INSERT INTO users
  (id, email, password)
VALUES
  (1, "vivian@vivian.com", "$2a$13$oCCafEIoJJZx/R31iGtOmuGULSIKKnHtytkpAlEYVMWBAuhkWx0Hu");