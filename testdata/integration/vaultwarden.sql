BEGIN TRANSACTION;
CREATE TABLE __diesel_schema_migrations (
       version VARCHAR(50) PRIMARY KEY NOT NULL,
       run_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO "__diesel_schema_migrations" VALUES('20180114171611','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180217205753','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180427155151','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180508161616','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180525232323','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180601112529','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180711181453','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180827172114','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180910111213','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20180919144557','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20181127152651','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20190526216651','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20191010083032','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20191117011009','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20200313205045','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20200409235005','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20200701214531','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20200802025025','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20201130224000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20201209173101','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20210311190243','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20210315163412','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20210430233251','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20210511205202','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20210701203140','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20210830193501','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20211024164321','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20220117234911','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20220302210038','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20220727110000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20221018170602','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230106151600','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230111205851','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230131222222','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230218125735','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230602200424','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230617200424','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230628133700','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230901170620','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230902212336','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230910133000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20230914133000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20231021221242','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20240112210182','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20240214140000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20240214170000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20240226170000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20240306170000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20240313','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20240605131359','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20240904091351','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20250109172300','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20250820120000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20260309005927','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20260425120000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20260505120000','2026-10-28 12:55:13');
INSERT INTO "__diesel_schema_migrations" VALUES('20260902120000','2026-10-28 12:55:13');
CREATE TABLE archives (
    user_uuid   CHAR(36) NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    cipher_uuid CHAR(36) NOT NULL REFERENCES ciphers (uuid) ON DELETE CASCADE,
    archived_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_uuid, cipher_uuid)
);
CREATE TABLE attachments (
  id          TEXT    NOT NULL PRIMARY KEY,
  cipher_uuid TEXT    NOT NULL REFERENCES ciphers (uuid),
  file_name   TEXT    NOT NULL,
  file_size   INTEGER NOT NULL

, akey TEXT);
CREATE TABLE "auth_requests" (
    uuid                        TEXT NOT NULL PRIMARY KEY,
    user_uuid                   TEXT NOT NULL,
    organization_uuid           TEXT,
    request_device_identifier   TEXT NOT NULL,
    device_type                 INTEGER NOT NULL,
    request_ip                  TEXT NOT NULL,
    response_device_id          TEXT,
    access_code                 TEXT NOT NULL,
    public_key                  TEXT NOT NULL,
    enc_key                     TEXT,
    master_password_hash        TEXT,
    approved                    BOOLEAN,
    creation_date               DATETIME NOT NULL,
    response_date               DATETIME,
    authentication_date         DATETIME,
    FOREIGN KEY (user_uuid) REFERENCES users (uuid),
    FOREIGN KEY (organization_uuid) REFERENCES organizations (uuid)
);
CREATE TABLE "ciphers"(
  uuid              TEXT     NOT NULL PRIMARY KEY,
  created_at        DATETIME NOT NULL,
  updated_at        DATETIME NOT NULL,
  user_uuid         TEXT     REFERENCES users(uuid),
  organization_uuid TEXT     REFERENCES organizations(uuid),
  atype             INTEGER  NOT NULL,
  name              TEXT     NOT NULL,
  notes             TEXT,
  fields            TEXT,
  data              TEXT     NOT NULL,
  password_history  TEXT,
  deleted_at        DATETIME
, reprompt INTEGER, "key" TEXT);
INSERT INTO "ciphers" VALUES('e75b7369-7540-40cc-9976-9f06d0328e7e','2026-10-28 12:55:13.840529820','2026-10-28 12:55:13.840690153','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,1,'2.Q2s252rT9lmdRpw3HokC7Q==|bTpRySB+0t2jvweL8DHCNQ==|hLi9QoB6KCPs26tMzQ8ndfoWIl3SuC6oTpv7CyBujhQ=',NULL,'[]','{"password":"2.Xj44oEGHTGxkMgKZCtG3zA==|uxoRLWr8t0hCSrV8Beovl+4HXGGp1KDGgqwNXD2Mpqg=|5BDxuu2a0Eqlijv383lBn5GjFe6wKkHYH1yHrb7R3CE=","totp":null,"uris":[{"match":null,"uri":"2.QM6axdLa5fPQUpJbCv5hjQ==|5MNBrh4T3Lkw14SbxNn6i8IUGG5uXfRTDL4etnbsAIU=|ioT727Vj211qvDsoFK5nB8YKJGxhjkkSmssUb3Jqwac="}],"username":"2.jeSRaIuKwtI8wzVYrGJQ3A==|FAQCNRtYTSfcU9WchEPunIUxDmi7TN4rwioPJ9JcxvY=|kObc9csK/4Kma87ySQL/n5KTmyf74NSKjvbZBekuUK4="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('3b37eaab-d857-4eee-889b-09fa56113b26','2026-10-28 12:55:13.842444110','2026-10-28 12:55:13.842591985','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,1,'2.3J/OBg/vFyBZ9leGOCTMIQ==|5PQ+O7oHOZIkA5tde1WXBQ==|sXUq7wUyofX1TA2YBaJ2u3jADI03i89hMyklFZ4sJCo=',NULL,'[{"name":"2.HoXzZTB81eOOtReVw1SBDg==|9CQLqYaHl+QKDQMbqx/seQ==|SOoYWA4ABEn0mQyE8yi0+24Y6Cz7KfpLdQmtpD0jjVE=","type":0,"value":"2.lC9TGlvqpVgPKruQ5a+cmg==|Zd+DMu+dvPWxqpImL86pMg==|pQsSnYNRQhgKiR6O8telpEXgrwI6cChoNgDgivYcnmk="},{"name":"2.6m+b989cXhS5Xd6FKCHeVw==|UqVtAnQXUDrDEt6rhvmEnQ==|hTsPw/6oPbJw3FhQzS1co+JfizPikmPZ7X8fSmHD20Q=","type":1,"value":"2.M2HAS1Xh7YyYMj6rS+VViQ==|Wp2gRYXGWXu3QV+v70bnxg==|RuIkBcLiMGSFU+K7i2i98cIOZI2n8GfgSaFRa1IT5dU="}]','{"password":"2.Kuz3tXVVnb9173otKaX7iw==|kdyEWAC/5xjU5PVBSBH+BA==|XFuUFv9SqwQ32qnpGzGZD83uSXpCmBYCmEvqWgbye84=","totp":null,"uris":[],"username":"2.S556DCSg0bxRmrc6SiEN0w==|L81n5PXAIpgxhqzqWIVOIg==|50twNkp+jDk/4ZP4yikiLZT1ZF41bUFIQGh37MSoWCM="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('2a817f14-6f2f-45a1-ae77-783e5f299f94','2026-10-28 12:55:13.844111526','2026-10-28 12:55:13.844209817','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,1,'2.j92kLiPX5r0LX8DUSMG7oQ==|wBy/AAaYa6SqsIY2hto4sw==|ejoa8edP8H8EJOroinSP3yme71liJ2dX+GW/0HufIns=',NULL,'[{"name":"2.DIPli8CZf320xPkuUgRcYQ==|kq6/QYJ0nmuJOW7Ga3TdKA==|MxUq9aKJNoB8MGnnc08T7vyM1zrD1xBm6rMbfI9NcrY=","type":0,"value":"2.TgxENmYQZ1272+YT52tEJA==|KZJlkUXM4PCor0IdJt4PFA==|IMnzXrGcsbpkeruRj4FYE19xR8IYGrqZyLOySBM2mrA="},{"name":"2.SYra9/SvAimDyorWCQnmVQ==|36EIZQFHdJffv6tQnl7Qbw==|InuBsOtFuq5nEeDIniiOXKzE8jItmiWmNAgwgUH59jM=","type":1,"value":"2.ZQLUs4gp40leZap2MEX4aQ==|op5GJWm2HlazVWSCJNC5hg==|pgldLOUAbwrcK9iX2RZKY8L2w15bsnE7jA7BnNV9lAY="}]','{"password":"2.qTv0K+BJDFwSnLDzrTCexg==|FjvOEmT1purCkX8WR0XskQ==|6+kl7jChGkIX+KbQsVS/scuYGpyeSZcEUXU6WxogvOo=","totp":null,"uris":[],"username":"2.zzUGyW13jmx56SbR06VQ1A==|2I70q3kT25HEZIOoN/Q8ig==|7AVyyhUEeXfqkX5bzI6suttySv+Eh30PB+H+fxKmAio="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('8c689210-eaa4-4530-aec8-51bf0e84c6e4','2026-10-28 12:55:13.845769691','2026-10-28 12:55:13.845877691','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,1,'2.PzcArizFJ3O1ukvPbVEIkg==|03ugYoQaAMxt+VGKp0MA69yW4EOlvAeQXp2nJ4fraig=|TIrn69c9Q+PgvkmMdfPwDUJRfwOu6DQk39reKrw2ol4=','2.qJv9LmoxJcS0vy/iWRtMAQ==|7cu1+iUuj+q4efJRBDSuocUEP5IxcOko9G87sk/dFAM=|dnQsVFb91heWoAVmbcOGIO+ALQxuDHGQMi0+976Cmnc=','[]','{"password":"2.Q7NAoabnqcR7vNj3CBC9rA==|nY6IocsrZv42bjeogUAHurDYBIv/cdINFbMUxNHRIq8=|tAsSaRs/Hlk+cxqBluUSjxl//1tVupe8pLmrCaRPgU4=","totp":null,"uris":[{"match":null,"uri":"2.mUQH0rgcsXATaCZrUiQWeg==|C6kf8U2wu8IKQpjGXFG+76rtMytvc8R//dhw85ZvRfA=|o/ZfCQ6CUYi/jbG6iT0xKj5BSDNQeN/Kf8Jqg5djC4s="}],"username":"2.xV//7NR+7BBOUk9vDEiZYw==|DjUUFZ+iJPQ1L1005vzvfr94pe21XttvzooezZNaV54=|ZG1k2ftsFo32N0Zk/bZagf0ht2PIJLY5WT7JGoxWQsk="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('32ff6c17-14e6-44b8-9509-a4b1d22342cd','2026-10-28 12:55:13.847301565','2026-10-28 12:55:13.847365773','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,1,'2.i/KUuri+Obs3tM9Gmn9mqA==|nBADUloHEG8wGsR5Ygpgbg==|vGsLYc8oWrxN0AyjrStl1cm9eySy06UEYdFXvjKACck=',NULL,'[]','{"password":"2.PWnXyaI4Xs85GcCS9bBGIA==|HmRV3ISuPnPqEVvC9dd/sQ==|0/o3DWWYsLRuJYBBmq6AQDE/8S6kAbxEkh4fDY3CgS0=","totp":null,"uris":[],"username":"2.vNNq5HysgrAvH4D/6Q1/ug==|8XaK0mumPW5YbzlAIgT7tA==|v5BqL94J29rerTclTB3MwFSDHLr212qi0wQq88ex07k="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('41f5c23a-1f3d-4094-9449-e484accc67f4','2026-10-28 12:55:13.848597273','2026-10-28 12:55:13.848691773','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,2,'2.ek1balc0VkJjBCtPOqrEpA==|OYQRnNJsHoL9Ec0nw2phEtUIBNPij/q3ZNoeonWYV8c=|cwB+bLG4Dt9ZEGwtMVBpIy025OJf+MBOPB6OsaIafUw=','2.x3LixzgYKAYXbzN3gyyZZQ==|BLwkKLrmkVR5Fyo0bujn4KNeomuXcgIA+Z+lZl2Ark4=|tk6ehI0XMuHSlo2xNuMo13ozcZ4/Hbw8FiawKGvSMfs=','[]','{"type":0}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('629d4403-edaf-4eb4-acef-e69ff6fea96e','2026-10-28 12:55:13.850234730','2026-10-28 12:55:13.850309521','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,3,'2.eBf5/tevMG0GDV9bBELV0A==|u5Bj4vtUkLR0LaTdFP/L9ahe0crjDlHbcO85/9AigLk=|okHrIX0jj13Og4m61fviEJ8KTaJrfMSjtLwa3A9sm3Q=',NULL,'[]','{"brand":"2.371zcxXpzu0OviGsJ+3Drw==|D6oEVGWPtfPSfo60uRwcaw==|uaVROsTJFyfW9vjYVbMThn/6bH0iu9ttDvKe1KKQgPg=","cardholderName":"2.tUWQC2e1zZGcq/D5F7O2kw==|aOpybjw7JJwR/2RT1BaThw==|QwPxjxEzxYPpVoZBEjzG1bpjjQ3ZWXNOreu7iRjtgBk=","code":"2.yrJEnWcRggMLrYTgRdKnJA==|uuE5TVJDSX90vfzE5hpQSw==|c6a9cW/Z/M/edKxRCuzW91v1l3bHsiN/ifKMD45joGI=","expMonth":"2.wC8KCRnvOfixosxv0l7brw==|kpHhbAr/T6tdkPy6wdq8aw==|cHhIyl8uHkLThwsi7cQP7HIOJz+5byYogIFzDWZJR0Q=","expYear":"2.BTBZIrrfJquzTohkiZ1eUw==|IlWpk1gvptR/truf9gXavw==|gYAGmQhYOsYdwXj0LnfRIp51gFbDHpl5I+tyVU7aLbc=","number":"2.VVg7m4inZXrnyZhm9KQkyA==|9Lbx0CAmylf/swEHpydYXwlr0cPpQnZssjiS6St3Yag=|EAzHoFE7wKsFD8me8Vq3igcJOiwm/szRwb+Uej0UbrM="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('b098bf6b-1323-4fb3-93b9-649cbcdf9680','2026-10-28 12:55:13.851638145','2026-10-28 12:55:13.851731145','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,4,'2.cI5i/VFbhJYAT5BWDz+fVQ==|4QOj4nv2PRgHqcyUuuEhkg==|hcsWrVwoXr2rpqRuzc9Zz7ojlUgt7Hx4dfQOK01uzpQ=',NULL,'[]','{"address1":"2.DkY8jMJQlHeDcf/fhishuQ==|Y4yqsKGjkvvZ96EuatuIQlfhKAlVTEJS53okZ3cSGWs=|yIUvd64gNn15aY+xqs1S3xo8Rk5GuE5Xui6UkndnpwA=","address2":null,"address3":null,"city":"2.dDNyOOVH2YyQ/Kv7uzqvQw==|9kTfocAtDedM14cIWNGfNg==|i/0LU/i4iRkU5WM1hv1ukqIWlsdr1/j3wUPALlmP+YA=","company":"2.qSDh/w4caD8plWsnWKFxvw==|yW1UK9EAKjYBdcfa5u9Yvg==|38MlT6qazgkQwsSWUm7IO2gRDjOM8YQoMQIENgKG3hc=","country":"2.UbD0X7GUCqh7Ff0QU2HvoA==|a6FyyKEsNpGGBOKCTBUV8A==|3PfaKXDlxjOUWwa4DbwcIWbS8+fpjMJgSSrk1ZXm5wo=","email":"2.zNe1K3KeIa+j9O8b0xLtiA==|c1f0Ul9dlZ/zVoylT/2us93snNh2ZkWfm9lylxE5J2o=|rQ/xHn5QTbikDm3p9979KgGjuiJLgkjCnQdx0ZCX+Ng=","firstName":"2.DFDYkME5FbL+ulDbjUjvEA==|47vJxUnB+CVTLzonLeLQFA==|NwaWcFWwt/fykoxsLFeW4uHoXsvu7JE1I/WYOwr6y+Y=","lastName":"2.2bIIRtPQMvyrRbyajZdF8g==|k/4iDpJhhwDs8hL0gZCosA==|/y/hGhS0fEKjvaWvpLi1zjgJxuGozAgFE0pzTTieCoY=","licenseNumber":null,"middleName":null,"passportNumber":null,"phone":null,"postalCode":"2.oV/tenfjgkPY5qY8ZJJ5XQ==|veoLC6ownO8mytYAlieLWQ==|Yy9fOXM95vFcaihorjAxNw/WMr9lVqY2Xlzd+dngFdc=","ssn":null,"state":null,"title":null,"username":"2.rEmsxWis8Rdndx3AoQOeWg==|Ar8wpoUs/FJV5HBHsyTuHQ==|m+7NSFhelggpSvweTk3HuKGqzkaGHPG4OmmkIcFD78Y="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('003b5ef1-478d-4ced-873a-991387dd5986','2026-10-28 12:55:13.853076478','2026-10-28 12:55:13.853154103','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,1,'2.7TMB/JrDL5VFsXnbM+B14g==|pfxpNy6JVZ1v7u9riF4iepC3fo+nFbBpRkg5lb+0+/s=|TlJjSDSXMNvRD+yn9pW8yZDCsueWJlwC4kHoqvjbzkw=',NULL,'[]','{"password":"2.D7VdlZdWWdzcAA0UJtGbmg==|k3wuQrPATCRD/iF7pm670sctGiNN++mlwlJ/FJVX72Y=|3kAwSDbLWjpeBDV4iX5Y2JdN0FATF6xf76On62uCe+c=","totp":"2.kY0INATJawinc6QN7RvDsw==|rS8nCzTZdK5J5dJY/O/yHGgk5mOtOmLJrR1jYa7OTmjNhzU5cO6oUBCn/iCaAEV2|wXIeuPjuHLXjv+j9DH+hdp1QPSjXCmbiWwrsRK+012g=","uris":[],"username":"2.HDf3FTUrz4wZS8cErdm2zw==|M3K7W1VJB5L7fuiwyjGfAagpdfrOyuF7+kOgnZeB3T8=|G2LI/WWNuIJk5HuqMw4xNuoxkVaD3OIJNQUqFs8qlCI="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('cdf8324e-baf0-4960-a7cf-6661b8584d9c','2026-10-28 12:55:13.854526393','2026-10-28 12:55:13.854578810','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,5,'2.P7NDOzTytuaR9K34EoqVzw==|Glfva7fzNAQIihekwbJ4sK1931Jn4j4uyXAKjIc4zjc=|pHJ9w0HJP+m1L/eIWBkOwRR7vbo5LxiU/Tog8HppPUg=',NULL,'[]','{"keyFingerprint":"2.mCtrA45k2mDpCTaaQeBD+A==|jVcqImQcMeV50YIogqbEvf0sYUhwv15Sb8YzxNePu5GWfetsUDxxla7ClmqHv0sGb1LBh269qYy6kLmJeZiyew==|116jFEl6ClwPuUB24O7uH/Gw/LHHZlkcQbjsovTfT4k=","privateKey":"2.IYOHKuyiRIzO9Wtsp30vEw==|+XqC1LpaqwPDK0qfXLPH/v+9gd/pL1bxiwTWZdDfLsvB8UxRliVzBdF/YxnxrEDVPD6OZ2XhvFvVB27RLjaRD/bAbg6XpOx0PrBiEEsVx90CXEvU9nhG6ZkwTimatOU4Fby4RhycGcjUHFl3rxSLe0gqOj+vFu8qBpooW9radDUVU3RPdjfZMFEcXtccnCSiWD3ej+nwW3rpIuj2FWgmoYZJzmwxAdd9zjiDLJMp6ukJiKYMIgZci9uj6DvjIjZ75w0OnAKfJgk9p1IlkDQaUEWpejvCMJRu3xwSoWf6xBRL5ka8VbwGjKPCv/ZG5s0kshferU3Q6cCx7/+2v82+/OEGcXbnu2LZweRFhugxDOOx4IMzZ/fvhUIC7zCudinNtDQnr1cCgOIn9ZkqerQ7zQIUx65hYXYF1VQfbPNqKTqImitblBHCxzaFnwxYIL+eUpTihBkzZGshozRc27nwFd0usAgrW52+V5k/fNcP1E4UArQ21BblCOPoyceV3Otv82U6KdcijHL9WzlHtA7I4g==|tA7ix8Z+ot74pnz908y9ObzP0wJQpfjJEakhdeUy1JM=","publicKey":"2.veCHDVetsci9HG4luK+Gxw==|1x2fHdQ7XldWEL8Oz2D4OIqSa5r6qqt9bV5vS8rJRCAwIxUvIT4PK9CqHx9hj3wJsQcFRFtnrwQpEG6RC5R4j+yMhJXdomXyRuKqFwpCKYFZ0heYRhQRQdKapn3/krg8|7KEggCXWCT4B/BE+cZ2ypDBTbb/PNAO5sRGyXItO738="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('ce3b779d-07c5-4f32-8956-8df6b411f56c','2026-10-28 12:55:13.855972351','2026-10-28 12:55:13.856037892','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,6,'2.4JhP0wKASY9a9CQWIaHkCg==|bZgRjyL5EXCOE+y0+7VH0ZvjttItnjFc/93hr5uVT58=|6g4JkWsB26fLvaPWohSZCKDHI+13uIvKXPzdAO07G9s=',NULL,'[]','{"accountNumber":"2./IF85d4ZH8GFOnc94WK1fw==|OscV1vCrh8ViJZrTvg2hpA==|9vnlMvOTjXtHoohBMmOmlnbzQhNccYSLd83JxYXPPVY=","accountType":"2.U5i2hRreF2leRbzE0FHLQg==|qT0Zw6iF0zxOKPtZWCQ55A==|9C3rhO4LQJ486BbgEADDTOTnMy/BItPcgnOg+sNNEJA=","bankContactPhone":"2.SQjONlFrFqGjhUsZKMa4eQ==|xyTtMySLpieH5JM57OWvmg==|yr+giZ+eCOGtC23tutwrAothFR4tvgnz4Wu1wei4vJE=","bankName":"2.FJtx1TNsZHONMLdld6DZCg==|Zm2I3lKw8ZjfF2m/M2rUxQ==|e5SgqR44YP0TcENnVqK2+ALor+1e8KqvGtbT4t6ahZA=","branchNumber":"2.UWKcvJwqM4HfRclGfe7vyw==|+OLvQigKxn20PQvpzIKfow==|z0+RDWY9V2CS+mOHJSkPcYAPkK9xPI6Tg8x69OCbu8s=","iban":"2.doJ8jcW0YmPzgdrTdqWuMg==|IGH/4vSrW/BIOhLzy8jmkQ==|Lr9yiQVugWSevBPufcjoiLh+kQTwy+jRqhIMtFSPsTg=","nameOnAccount":"2./dL8Se4VJj8J17jl7dPTWA==|hkdjb3aDluddz2sLyQMgKw==|dNmR+hAGGaW72GqvIJO1OAFwhauozUznPR0/X7SU7eE=","pin":"2.5kaqUai+7y0QzW+H3w8iMA==|3ybHpH488a0LFde04YbI5w==|wpAfcB2ffGq2UGjiBAI7CekBH59SRh3t+c/8JNaN5XA=","routingNumber":"2.kvKfoX3HByHGnna1BypAlw==|tedO/5d6jQpMVuePYsATZg==|TQCsAPFKBrHsY/B8d6ADbAUAyBdGhLzWQjs1pwYG/ds=","swiftCode":"2.7Xpjj9nPRzC/nsWfkWCk9w==|UKWRhC1uss+pADiRQMe85A==|vvwIMDUXmJhv5VduqCl9PBHYdd3tkC58KxkFXJ4ZKqE="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('8a8a4c9f-b434-472f-b2b8-2a1f07fb9741','2026-10-28 12:55:13.857389391','2026-10-28 12:55:13.857454558','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,7,'2.Q+FlhA51LAyBGemfJwuwaw==|f1dE+FJLi/6PYy5KqW6JI/iK+UPq/Msg9zZLubI+OFM=|twtl4bRMGBbgzfruS5ashBV6QfEIDxWrSoJmdfHJQXM=',NULL,'[]','{"dateOfBirth":"2.6JTQbqmlXV2/wTKnflHjSA==|dIoDbzFkiL25iXvqV5iz9A==|vSpMfSnZytDJOTYPU0SImThD2Z15CakpsA62+DV9dLs=","expirationDate":"2.OChhBXhcnDicEv7SbAVG/A==|yqKnSQgob950K22Z7nQW8A==|L9dSOBEmtfEVuyRz07/xzVZapKKh0PX9OUPXe9MBPBE=","firstName":"2.iv/oK2FjpBKxmqYnQkzouQ==|5p4ntwl3vNMCGTO4dhPIHw==|M2Pxz1JYSwdgPM8F65qEIG21cu/qyxOa3db+94f37Lo=","issueDate":"2.1Am9CYU7KdYH2+sFcY0u8w==|FgV/PIfhbYiYbVy7+yFZoQ==|Lu1qApuya1dJbV7y36wcZir1EFejsLxXIHtfnzUWFPo=","issuingAuthority":"2.Tukuz9guF11dDGodPZXrgg==|FH/q1iRbndWKMlwb6k7GZQ==|gRmaFo68CHf5WygDzzg7aUdxk3psKlKyeNsYekabm6Q=","issuingCountry":"2.nqGr3+uWqrrzzKMi17lTWA==|JZ2criPt3FIX8XiytKhpTQ==|M0ZOeqVyH9zsacIjXbI2PGU3xPCUb5BYMQdUPrwje+U=","issuingState":"2.0wvnnWcEWrQiwGbNhO/47Q==|EcU3QcFq+WziDbPP7PTkRw==|X/MgF1zr3koYdNCI47124qkS+LmQHKoViSkCPY93h2k=","lastName":"2.GkIiZuZblk0HMiZBnP21ng==|OQ6AiI1T8VC2h0Iwe5SICg==|8Pyg43mzkQrTztbsHUo8b10rFcHPFjpNC1olPocMG5U=","licenseClass":"2.zd3DvdJccUwfAm8rOb3LSg==|v0deh0L7QR6V/Gm6IjIUKg==|h0uw79hGdCCxH+8wnS9cxHzKqf5pCG5Ccj9ugBbmncA=","licenseNumber":"2./5JnyoYMW0wji8pjisk4Sw==|T+ABCmGpVfcnqYEi/m+rWg==|EYoiraQrA7Pm7kRQPHKmEN64BPVtioYKqWu7hty1T40=","middleName":"2.Zd7Ycl/7u3bcql03GNG7Pg==|K5Odzqp18u5USTLMvXuKHg==|S1iIYFj7Q5fkWutTsn5+SorGRwzsiHYYj/WJ12VFOOk="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('fed7342d-aa12-4209-9939-2cb0c90c56b4','2026-10-28 12:55:13.858857640','2026-10-28 12:55:13.858917682','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,8,'2.blwEpNyjTvZ48+FDLNjY5w==|HxsAH7TsapbE32ov6Mwtjw==|0Pj8dTOxp9gvhTairiKD+z4LUb0hvZK+jN7H1uj6Bp0=',NULL,'[]','{"birthPlace":"2.nR1kCDhOSpQ9GLjXkNpMeg==|uS2i84e3wrLLyhAPiTcJ1w==|DA3DdadKqO6gvls6boEx8Rgi9AWOaAcoKjh5ee048Qw=","dateOfBirth":"2.S06rUBKbXJdZSiw1Y5IlAQ==|TsHfPI7MVy0eWcGr06R8iA==|To9guNEF0BPqO2+2OWYqOBjaMRf6KzSIY93O/MWRy50=","expirationDate":"2.wkAFX/9VcStHYGUxg30EVQ==|CoB8A7OcEPW/0VRgMmj8Yw==|Ot7Y03bufRLKtw6RIrDgC1z4aAiIYvnHqxXWhedPez8=","givenName":"2.HbnSeh2OQbd7QbsWNJBKTA==|YypNKW2VDQP8YKfj9daxzw==|jjhtx/6hg4Sjw2PkWtSt7LuCR9Jhpia4gQGVkwBuOns=","issueDate":"2.Oa42dfSDeRSsmWp3qwIGvQ==|8pGcfOe6hr1nV1dwWwpTqw==|YDXkBQQlnGI5Gqeq9x1947t7gTMZZDKPHWVuW6JbM2E=","issuingAuthority":"2.gUKsL1Izu/zecEm7zv2n/A==|vamfXdXcyoChEu+hWflFaA==|NYxvXBw1A10aNGZ3G0wH7vfSC8qfC/LCoYiQoTug4Co=","issuingCountry":"2./W3XCTOSacO060LKdUXFBA==|0/wq72kdwATKeYyp+xtPEA==|DQig1ffBkA8cyEq32pKrDUxIiQPkJeWEhbNIj2Ja88o=","nationalIdentificationNumber":"2.LrM+GiAOGATIw++Xj+katQ==|iJLzr/S/JlnwGsa+htK5Jg==|PpRA7Giukg/LLscH5jik97LJnKvFUTJ4wPgnALzx9J4=","nationality":"2.olum1/QfKDnqMt7UV1Y1Lw==|/ryorCwSBNIgA+mNESu+Kw==|uuF0Nh20xBFHdjkaWJo3waksgnZ/x0Ht4r+PvQ0AjBw=","passportNumber":"2.wvXSP+3vKouMlBEUo+0EPg==|daVeoBgphFvbt1THBPVGsQ==|pvBwDrS2+cYWs4jgl6UIW6pUfyPjhX+qRBwt1VWzk+4=","passportType":"2.UF175/DUdnp9OFnYhUUVvg==|yL4JvQCi00Os6QKTyVygRw==|qs6BmfSurv/9Mp243Z+zkUDHkIOTq1v+FjxkXVMddAI=","sex":"2.3UtaX5bdQ5s3RmquY6DTHg==|7mPISt/FkFHHMEBEI/fkTg==|A6yY45JoUNwdGTS1wLGyLn+MWS/1xTasYvmBEKHs2D8=","surname":"2.Ld6i+23GDVMPDiEjoe6k+g==|cVWYOb9Kl5EgWldSXJGjQA==|VlospjncrHR2QSG6zNoxaAUoYavJBPmWKm0yHwftr48="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('1280f355-be69-4250-8697-749dff44a731','2026-10-28 12:55:13.860206681','2026-10-28 12:55:13.860299764','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,1,'2.ZixoWP5NNfc+oQxa9iyLYQ==|+QgX8jIa2bKclke7uQ5a0e/eWXJKBbslXxd8PSBZ1f8=|pYBtyrcEveYqmtUj3CMTy+uZz0Htr9DThvKc/vYGCYk=','2.UQeVCeuFscVC1r7aBr8tVg==|ryXEvRZcx9bfYGmYKc00tdaWKGbSYRdlvqu7t1AsrSA=|JT+fEhrfGHnsxRxNG498HEtn5AQ9U2iey+I4GAyyRd0=','[]','{"password":"2.PHt9137fJ4ArMZd7IPRB5A==|2O9VKnaz85e/eAk14aLotX1p1hkOTwT/T4jb3i/r4b8=|9IAqDy3QJWB914ZfOWloGBEU4ZAuI7zEXNAG8qpS8gs=","totp":"2.Az2vmdcNv2x6STthb4VcCw==|Wxg7OFoYsvQqKrkuTsQT5r4JuGGR+uaoZVPv5sGTqiA=|6LZqtgCZMDy4BbVfS6ZtabEso2gdP0Nrs4s1K7S+TJg=","uris":[{"match":null,"uri":"2.BidrSIWV9rov6Lq8/o5GXw==|/Lt97G4kP9C9XZzCYxYOo5CezcqP7Si64vT3v0aQM/Y=|b46kQFtqZ9ft+sVkAczi1cTuMuEdTv2fZr56ltQrwyg="}],"username":"2.nlkM8OeNy0ZOs22IVmkKIw==|rAfMSRiYlcrecyDK9I3Uv3QmW9VqE7WtiD0E1UfXnPU=|Vm0VSWpmdAqnzm1KadZKsUehqskl0ssm7Yq2oro333w="}',NULL,NULL,1,NULL);
INSERT INTO "ciphers" VALUES('d731110e-1aae-4ee0-99ef-f59912c9ba71','2026-10-28 12:55:13.862053096','2026-10-28 12:55:13.862143805','d97af672-5183-4c87-bc85-97bae473c4fc',NULL,1,'2.10F7Po/wOcWib6RyIbuZTg==|dR/N+Pt6Vl1+zcamI8KAluUc+JSMcju+LZwFVd3WsAN0d/ah73bibM73+lAkhHAH|VLb51scmHf7HS9zmNGJMf2tQ/UqHofjlJLvDo+RRC0c=','2.ygu/lA8YeIFAMx8IvynSnA==|qBbsQRjMtpx6rMx61/8MJrqeUuGHT6AU5HVFEmFPJNE=|eFSUmbAbrvUUYsB8/nxWsI5rP9TIUE1QAevbLU4ZE5s=','[]','{"password":"2.KcKWJSzORDNbrCwtQxDgRA==|0CiG4mY6bYDZShIpa70HD+BCuJnWGV1L1Pu8izZlqkA=|5M/U2PoRBSUGNlqiz+xb6hCq/cRw+O5vdQ335aBWuQM=","totp":"2.JAui75zIb7JB6OQ68ByjfA==|iyEPyWLVtkJoFwRDdFzBeNA9awnRCLwlPTe+HF6a7BM=|E/d0/EBKmAGMwX9OUg9xbVuSa2KcJqVDHlDOsMbZd6c=","uris":[{"match":null,"uri":"2.4wsysCRzAadqlMZY1FyGog==|1Mpqmun0enmdIUIEymTDYxzf1TB9V3pYR62RhQW6GtM=|Fa2WkJBcsahY9dZ6vyzMbs5kotBp82TIMgNt6EIcx+Q="}],"username":"2.GT3tnwwW/SZ4HyubemXY1Q==|9X7gVGYlZWISuR/9kh93VDhZxkjGej2BkNO1dQ+g1SI=|Hco7HZURi6okcLtRogNpdRuHOCW+I2h25qTbITjbpn8="}',NULL,NULL,1,NULL);
INSERT INTO "ciphers" VALUES('ff1256eb-e1d5-43e5-b2e5-9f217a169628','2026-10-28 12:55:14.275946680','2026-10-28 12:55:14.276685180',NULL,'5f25abf9-a963-4086-81b4-48e48ba9927f',1,'2.zoQnxH8o+QASHGtmi5NJ7w==|TbckYNJt9s8gz/VYBtzzXo9t6zG0avKhbUVhTai1Gjs=|SxdC3kvjBPLaGU+MkYbmkcZHlgtqdvfqF32s81Jmepw=','2.sMueZoAscO6gq4rFWcPBJw==|7NH6oI8SvSRLZ9HfYfyTlvOuNsRh75uPP51LkXc305o=|USF8FeV6POwv4cl0XidriBM7YWC3NZa+67y8uTvD+Ko=','[]','{"password":"2.wLvC6aTf+p6SL0XmA1sZrQ==|3Xo8W/G1DY46zTyzXr7T7Q==|nrJkl8EzZKvNm0/zmh6fOv1VJRLzJ9fr/OJde/2duoE=","username":"2.PwVBzWN/I5Qixa4y6Ze4AA==|t5eagLrxJgpRZFcaHc/wQQ==|Jv7tE7dBOspROPTrcEgnf2Oeg6BoCbIN+zNsxgHxLEg="}',NULL,NULL,0,'2.OF+42KcGjpmtSiUicKYRnQ==|QqFV//hf6wtGS7bTxr9jzjtLMLz6Ces6FvnGR/fyj1kvaX84PXRawHrqtNJK3weUt9dkkf3x3wI8FCmm7b04YUUHtgfS3C1Var3rKeSfLtw=|qEFBNPUWors/H7bS0fkWXVNIKExrXM8YJKTaS6/Wgq0=');
INSERT INTO "ciphers" VALUES('86cb784e-3f62-4c74-84d1-edf735f406fd','2026-10-28 12:55:14.278000721','2026-10-28 12:55:14.278394095',NULL,'5f25abf9-a963-4086-81b4-48e48ba9927f',1,'2.nFr0UjdAG9oM0sOUKOiXkw==|o5VgI113Gm49phu1BqAGhQAMKW7RpATqpY8jBpH63Zc=|9f1vIR+FdovWI+sIzdeg5hczpBzL9sEgLPq5JOHMJbA=','2.fhnQi1nQVAHmE/rn9dcL/A==|Tr/z2jh2ZR7LP9i9IXiqrWjzu23adth32QTbjQbRAkU=|DMGQ66tNySswjOlPKIalpCazVCDS1BxGumcasQaUHIo=','[]','{"password":"2.l59ArXSjDhFB3G0tqOG0ng==|+B9Y6XMJM0MCZX934bpqR0e7xx01EYb+Bsp25XPB9u0=|CSzAn7MOjDI3TBEjdLMRxmQJT2WOmShLCi5PJDHDhic=","username":"2.CrZpyzJ6R67uw4PWu4jlow==|3x/O9/W7ccaem/nRLGKYPQ==|arW1gB7p9Cte65XQ6JfDI5Ic/vwlZrn0bvkPuxqTHAw="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('bae3fe1d-5848-473c-a567-d820523b62e5','2026-10-28 12:55:14.279637761','2026-10-28 12:55:14.280130802',NULL,'5f25abf9-a963-4086-81b4-48e48ba9927f',1,'2.aT9KbxNNv5kkwtMUWbt15A==|dXWO8W4IlyGD2JMRBz9L7WycZbkueHZngvNxSbpRUzE=|Rq1pLJAVFspG1SLUdbzaIxckqhbEBqLR7/vDibwLgtk=','2.3wsmgrH5glct33iKQr9yIA==|J4EhaK5AnuiFXgBS0dCFQbcJkLyCqlH3qVXTpY5adxs=|TKkohoxOZ13uuVrcPlbYO9ZWsfw3cedlCA0kpB+WOvI=','[]','{"password":"2.lNkixeuv3MOXynJI/LprjA==|5fEjH9Insn9sdANgskvxMyAPx998XDT+ZySMa0jHERk=|TonGdHiT3iUoO28oTmtYTubl+W6uNH9fRt4HcXo7ocQ=","username":"2.Fj79IYWLjXGhzRm/cdH3RQ==|dlyR1seAzoQM+WPoYVCxgg==|T+XZbwvuaFx2jMjxD3memIy9aBwLW+Iiiov7j0X0xXo="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('0c589ada-afac-4c36-a892-72c96f96fa6f','2026-10-28 12:55:14.289858712','2026-10-28 12:55:14.290480462',NULL,'e6e8aadc-2d28-4fc5-b1da-d02b97ab2f17',1,'2.OyFqWJ1wA//Mmhh5HVE+DA==|x+28n8svG1t3UwH/OSLbWRGxxOPSodYeYWmh+Y4kTJw=|xRBncX+1GP8LxwzLHEuy/lsUAPPrQfyr4HXLM0CKHoI=','2.Iijmn2xM1+2PzKhHgRkq5A==|c9pUCcn0GE50tUCHA3xlUcZiz4FlqkqGKh580t5DnaQ=|8mvRTuyWMBLO2lphyWZ63Zjc6ZB7pXzPdddMxC24AeA=','[]','{"password":"2.h3UzyHiWRmFDHjoOJ8EhKQ==|pYx++XXZv4UvnYsgaz68qw==|4aLivCHb9+oINDOZAMtt3wPwRygyzmf+6QCYfcEW6pE=","username":"2.2S8gyOG2cWhwpCNbWfsFpA==|L77M9gHPj/gpkhBdQ8l9GQ==|BHRTQo479BKomfgyB6AqkKYaBVdPNgkHDC/wbKOBHPc="}',NULL,NULL,0,'2.5jjK1pM6yu66UQ+uA3W7fw==|Eg882gCbf4fH/fFbML294/grk5bfVzOQVTGaUQ56JXX2+BqM6O5ilw+0tBOnBd76FmDh4oAacPrrg0JQvJNrPxqG4mOXM1oFJnvjgMViHcQ=|qBB20mugeA+vb8XQSjGU0zlUlxJ23F+o0VuhH34Gmvw=');
INSERT INTO "ciphers" VALUES('a1d38765-4f06-41bd-9cd8-f1caec8363c2','2026-10-28 12:55:14.291914127','2026-10-28 12:55:14.292347169',NULL,'e6e8aadc-2d28-4fc5-b1da-d02b97ab2f17',1,'2.tzPfRD0tsdOEYt6Bz5XfPQ==|7LC+gXCjHGnSUx7Abp5DW9HqIfSX8iJpi6NpsiJFEjg=|BTW4Hczp0n42l9Y9+ePFpJXAhu/qaRtBjmbxEoJePZY=','2.twdn0ueP5CdiM/7xbfQwTA==|req50cVczilzgsU01nx+eiXGmAi269sfKvIFg/BcgeQ=|6fwElqyQg8+3d04gG7wWdW0+MCdNvmy4fbzpgCKXuOg=','[]','{"password":"2.7mawsRmp8OKbaaZqXb7kyw==|+6+o3FwgxjlyQ1oy4wJ4dymnEukKysbyOjztcYC6MdU=|9XOBHIPvwTTJ0oTWnaZbjOLjDfn5M9fNiExoWoCsexU=","username":"2.uTkEYdP4YZer/Chl52Rqmg==|7/+F+aD+o+wtPbh7zheazw==|OXVZ/R+TMKxTUmC4go9xNQiS8s6rQDbKZiBRBtV+lxE="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('2b7a8129-6149-4dac-ac58-9028cc3b6247','2026-10-28 12:55:14.293547293','2026-10-28 12:55:14.293952084',NULL,'e6e8aadc-2d28-4fc5-b1da-d02b97ab2f17',1,'2.ZLtsHkRtS2WRaDduvrIIMg==|EdpQvkhufVNgp31YA8frTuN11Ha6imnyb+9vRGGaqL8=|amw/86WdTO55XoefLZRtqVz/z60E6gsryUcZhaVUg2s=','2.n2uJOe0k5BqvcniuN0vyyA==|PxxezzqzHFcrcEbS/RlNRID1X6tJ3Yl0bVG5FoPy7nk=|J8oK1xGLw2eMlgUsjC2Bpm7W5aCAGHC0m1S2g1oPvz4=','[]','{"password":"2.9a+rSHq2emDraTH6mL8pbA==|KWRb/P5ZOC/DKDPRu25Gto/6qz39zKtSspM3OZisRdA=|jfdcMBIxIgnoX91lycOSW+1p4VWcvyDS0gVL4Tvl278=","username":"2.4WHhAWBKnDT+zO7HbTSEFA==|o5oQb86GORDULRlS0a0o3g==|WJQx6W4K3WUuNCUYUXQipAjls06lhjc1xBpOtWX6kNc="}',NULL,NULL,0,NULL);
CREATE TABLE ciphers_collections (
  cipher_uuid       TEXT NOT NULL REFERENCES ciphers (uuid),
  collection_uuid TEXT NOT NULL REFERENCES collections (uuid),
  PRIMARY KEY (cipher_uuid, collection_uuid)
);
INSERT INTO "ciphers_collections" VALUES('ff1256eb-e1d5-43e5-b2e5-9f217a169628','6a12ea69-7e50-4531-8a05-308947a8e99e');
INSERT INTO "ciphers_collections" VALUES('ff1256eb-e1d5-43e5-b2e5-9f217a169628','4f898ee1-1353-4906-92bc-04397f9e8f4a');
INSERT INTO "ciphers_collections" VALUES('86cb784e-3f62-4c74-84d1-edf735f406fd','6a12ea69-7e50-4531-8a05-308947a8e99e');
INSERT INTO "ciphers_collections" VALUES('bae3fe1d-5848-473c-a567-d820523b62e5','4f898ee1-1353-4906-92bc-04397f9e8f4a');
INSERT INTO "ciphers_collections" VALUES('0c589ada-afac-4c36-a892-72c96f96fa6f','93e7e5fb-5d47-4634-8fd6-aaddb9293020');
INSERT INTO "ciphers_collections" VALUES('0c589ada-afac-4c36-a892-72c96f96fa6f','a14686ec-8600-4abb-a160-72ef4fb01e7f');
INSERT INTO "ciphers_collections" VALUES('a1d38765-4f06-41bd-9cd8-f1caec8363c2','93e7e5fb-5d47-4634-8fd6-aaddb9293020');
INSERT INTO "ciphers_collections" VALUES('2b7a8129-6149-4dac-ac58-9028cc3b6247','a14686ec-8600-4abb-a160-72ef4fb01e7f');
CREATE TABLE collections (
  uuid     TEXT NOT NULL PRIMARY KEY,
  org_uuid TEXT NOT NULL REFERENCES organizations (uuid),
  name     TEXT NOT NULL
, external_id TEXT);
INSERT INTO "collections" VALUES('6a12ea69-7e50-4531-8a05-308947a8e99e','5f25abf9-a963-4086-81b4-48e48ba9927f','2.JWXSbBxWSZtTsZA+utQ4tQ==|nre9sG8jlYOcB5OBJYNhSg==|y7+M4z/i6Gn6kygjsBn5nb6cUsuAt6T+kPoc1ymIDEk=',NULL);
INSERT INTO "collections" VALUES('4f898ee1-1353-4906-92bc-04397f9e8f4a','5f25abf9-a963-4086-81b4-48e48ba9927f','2.KUVpI9idvrm42EIgqR0/mA==|4Q9fH61ocAqRdhoLlMxncQ==|bSSNQjaaz7Ks+wgDDDsEMNC2UMg0bLFiq0xzySh/6P4=',NULL);
INSERT INTO "collections" VALUES('93e7e5fb-5d47-4634-8fd6-aaddb9293020','e6e8aadc-2d28-4fc5-b1da-d02b97ab2f17','2.rWemXuyUmD5ocZ7a639tlg==|/mfQVUQBbR3w/wKoR2RR6A==|mKs6d/x++69I3PO56Du9LCKZs+B2gh/D/NrF7S1BdPs=',NULL);
INSERT INTO "collections" VALUES('a14686ec-8600-4abb-a160-72ef4fb01e7f','e6e8aadc-2d28-4fc5-b1da-d02b97ab2f17','2.OTglpgB325Su5CsfI2zIPg==|oqe0sYa1Q3UfyLLq/+L5/Q==|bp0b46/zUuPnoyjojy+lav/Zv2aK/gVMlS9Vv+4gfGc=',NULL);
CREATE TABLE collections_groups (
  collections_uuid                  TEXT NOT NULL REFERENCES collections (uuid),
  groups_uuid                       TEXT NOT NULL REFERENCES groups (uuid),
  read_only                         BOOLEAN NOT NULL,
  hide_passwords                    BOOLEAN NOT NULL, manage BOOLEAN NOT NULL DEFAULT 0,
  UNIQUE (collections_uuid, groups_uuid)
);
CREATE TABLE "devices" (
	uuid	TEXT NOT NULL,
	created_at	DATETIME NOT NULL,
	updated_at	DATETIME NOT NULL,
	user_uuid	TEXT NOT NULL,
	name	TEXT NOT NULL,
	atype	INTEGER NOT NULL,
	push_token	TEXT,
	refresh_token	TEXT NOT NULL,
	twofactor_remember	TEXT, push_uuid TEXT,
	PRIMARY KEY(uuid, user_uuid),
	FOREIGN KEY(user_uuid) REFERENCES users(uuid)
);
INSERT INTO "devices" VALUES('30000000-0000-4000-8000-000000000001','2026-10-28 12:55:13.830832410','2026-10-28 12:55:14.712140998','d97af672-5183-4c87-bc85-97bae473c4fc','bwb-integration',8,NULL,'ie34MVZGC9NddZBCUPq00wOuY0oU0Vo1jWbAx--XNhPDbvubbwEX03pexjFs-8Ezvt7BhZSBUftUkI8tkODUTQ==',NULL,'844a165a-8ef2-4482-b221-ce4a97917ad6');
INSERT INTO "devices" VALUES('30000000-0000-4000-8000-000000000001','2026-10-28 12:55:14.938279714','2026-10-28 12:55:14.938473423','9bdb25ab-07eb-4882-a1c7-462666e4963b','bwb-integration',8,NULL,'upllxaKJABu8lojZtC6sWfEQv5BPVIdhRvHyAZalFp8yyjn-WGZZ86UuytQ1B7efmbcn9AqwwqbQoPxkFdEYLg==',NULL,'37f106da-4ade-43f5-8b20-02fdea55e6b5');
CREATE TABLE emergency_access (
  uuid                      TEXT     NOT NULL PRIMARY KEY,
  grantor_uuid              TEXT     REFERENCES users (uuid),
  grantee_uuid              TEXT     REFERENCES users (uuid),
  email                     TEXT,
  key_encrypted             TEXT,
  atype                     INTEGER  NOT NULL,
  status                    INTEGER  NOT NULL,
  wait_time_days            INTEGER  NOT NULL,
  recovery_initiated_at     DATETIME,
  last_notification_at      DATETIME,
  updated_at                DATETIME NOT NULL,
  created_at                DATETIME NOT NULL
);
CREATE TABLE event (
  uuid               TEXT        NOT NULL PRIMARY KEY,
  event_type         INTEGER     NOT NULL,
  user_uuid          TEXT,
  org_uuid           TEXT,
  cipher_uuid        TEXT,
  collection_uuid    TEXT,
  group_uuid         TEXT,
  org_user_uuid      TEXT,
  act_user_uuid      TEXT,
  device_type        INTEGER,
  ip_address         TEXT,
  event_date         DATETIME    NOT NULL,
  policy_uuid        TEXT,
  provider_uuid      TEXT,
  provider_user_uuid TEXT,
  provider_org_uuid  TEXT,
  UNIQUE (uuid)
);
CREATE TABLE favorites (
  user_uuid   TEXT NOT NULL REFERENCES users(uuid),
  cipher_uuid TEXT NOT NULL REFERENCES ciphers(uuid),

  PRIMARY KEY (user_uuid, cipher_uuid)
);
INSERT INTO "favorites" VALUES('d97af672-5183-4c87-bc85-97bae473c4fc','e75b7369-7540-40cc-9976-9f06d0328e7e');
INSERT INTO "favorites" VALUES('d97af672-5183-4c87-bc85-97bae473c4fc','1280f355-be69-4250-8697-749dff44a731');
INSERT INTO "favorites" VALUES('d97af672-5183-4c87-bc85-97bae473c4fc','d731110e-1aae-4ee0-99ef-f59912c9ba71');
CREATE TABLE folders (
  uuid       TEXT     NOT NULL PRIMARY KEY,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  user_uuid  TEXT     NOT NULL REFERENCES users (uuid),
  name       TEXT     NOT NULL
);
INSERT INTO "folders" VALUES('70dec315-6617-4c7e-99e8-d78a5879563b','2026-10-28 12:55:13.837826405','2026-10-28 12:55:13.837908363','d97af672-5183-4c87-bc85-97bae473c4fc','2.9q3j6RkjCqsR+olyWWVowg==|zBwKMcVoe882CoV54BiKQQ==|iImQJ/yI4uBz7aCMXPZsIrRc3RNGDV0IN6WPKTswq0E=');
INSERT INTO "folders" VALUES('546679f1-3295-4316-9b9c-3e750b01262b','2026-10-28 12:55:13.839070946','2026-10-28 12:55:13.839130196','d97af672-5183-4c87-bc85-97bae473c4fc','2.8e6UNiEGH2mYkXS6Dbr3/w==|m00njpuPO0fWBn+uMylUjA==|VdpuF9ueFgO1RCIpHD+gbCH49vCZ/ZHr+2OgcljX3jA=');
CREATE TABLE folders_ciphers (
  cipher_uuid TEXT NOT NULL REFERENCES ciphers (uuid),
  folder_uuid TEXT NOT NULL REFERENCES folders (uuid),

  PRIMARY KEY (cipher_uuid, folder_uuid)
);
INSERT INTO "folders_ciphers" VALUES('e75b7369-7540-40cc-9976-9f06d0328e7e','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('3b37eaab-d857-4eee-889b-09fa56113b26','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('2a817f14-6f2f-45a1-ae77-783e5f299f94','546679f1-3295-4316-9b9c-3e750b01262b');
INSERT INTO "folders_ciphers" VALUES('8c689210-eaa4-4530-aec8-51bf0e84c6e4','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('41f5c23a-1f3d-4094-9449-e484accc67f4','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('629d4403-edaf-4eb4-acef-e69ff6fea96e','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('b098bf6b-1323-4fb3-93b9-649cbcdf9680','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('003b5ef1-478d-4ced-873a-991387dd5986','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('1280f355-be69-4250-8697-749dff44a731','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('d731110e-1aae-4ee0-99ef-f59912c9ba71','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('ff1256eb-e1d5-43e5-b2e5-9f217a169628','70dec315-6617-4c7e-99e8-d78a5879563b');
INSERT INTO "folders_ciphers" VALUES('0c589ada-afac-4c36-a892-72c96f96fa6f','70dec315-6617-4c7e-99e8-d78a5879563b');
CREATE TABLE groups (
  uuid                              TEXT NOT NULL PRIMARY KEY,
  organizations_uuid                TEXT NOT NULL REFERENCES organizations (uuid),
  name                              TEXT NOT NULL,
  access_all                        BOOLEAN NOT NULL,
  external_id                       TEXT NULL,
  creation_date                     TIMESTAMP NOT NULL,
  revision_date                     TIMESTAMP NOT NULL
);
CREATE TABLE groups_users (
  groups_uuid                       TEXT NOT NULL REFERENCES groups (uuid),
  users_organizations_uuid          TEXT NOT NULL REFERENCES users_organizations (uuid),
  UNIQUE (groups_uuid, users_organizations_uuid)
);
CREATE TABLE invitations (
    email   TEXT NOT NULL PRIMARY KEY
);
CREATE TABLE org_policies (
  uuid      TEXT     NOT NULL PRIMARY KEY,
  org_uuid  TEXT     NOT NULL REFERENCES organizations (uuid),
  atype     INTEGER  NOT NULL,
  enabled   BOOLEAN  NOT NULL,
  data      TEXT     NOT NULL,

  UNIQUE (org_uuid, atype)
);
CREATE TABLE organization_api_key (
	uuid            TEXT NOT NULL,
    org_uuid	    TEXT NOT NULL,
    atype           INTEGER NOT NULL,
    api_key         TEXT NOT NULL,
	revision_date   DATETIME NOT NULL,
	PRIMARY KEY(uuid, org_uuid),
	FOREIGN KEY(org_uuid) REFERENCES organizations(uuid)
);
CREATE TABLE organizations (
  uuid          TEXT NOT NULL PRIMARY KEY,
  name          TEXT NOT NULL,
  billing_email TEXT NOT NULL
, private_key TEXT, public_key TEXT);
INSERT INTO "organizations" VALUES('5f25abf9-a963-4086-81b4-48e48ba9927f','DEMO Engineering','bwb-integration@example.invalid',NULL,NULL);
INSERT INTO "organizations" VALUES('e6e8aadc-2d28-4fc5-b1da-d02b97ab2f17','DEMO Research','bwb-integration@example.invalid',NULL,NULL);
CREATE TABLE sends (
  uuid              TEXT NOT NULL   PRIMARY KEY,
  user_uuid         TEXT            REFERENCES users (uuid),
  organization_uuid TEXT            REFERENCES organizations (uuid),

  name              TEXT    NOT NULL,
  notes             TEXT,

  atype             INTEGER NOT NULL,
  data              TEXT    NOT NULL,
  akey               TEXT    NOT NULL,
  password_hash     BLOB,
  password_salt     BLOB,
  password_iter     INTEGER,

  max_access_count  INTEGER,
  access_count      INTEGER NOT NULL,

  creation_date     DATETIME NOT NULL,
  revision_date     DATETIME NOT NULL,
  expiration_date   DATETIME,
  deletion_date     DATETIME NOT NULL,

  disabled          BOOLEAN NOT NULL
, hide_email BOOLEAN);
CREATE TABLE sso_auth (
    state               TEXT NOT NULL PRIMARY KEY,
    client_challenge    TEXT NOT NULL,
    nonce               TEXT NOT NULL,
    redirect_uri        TEXT NOT NULL,
    code_response       TEXT,
    auth_response       TEXT,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
, binding_hash TEXT, code_response_error TEXT);
CREATE TABLE sso_users (
  user_uuid           CHAR(36) NOT NULL PRIMARY KEY,
  identifier          TEXT NOT NULL UNIQUE,
  created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY(user_uuid) REFERENCES users(uuid) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE TABLE twofactor (
  uuid      TEXT     NOT NULL PRIMARY KEY,
  user_uuid TEXT     NOT NULL REFERENCES users (uuid),
  atype      INTEGER  NOT NULL,
  enabled   BOOLEAN  NOT NULL,
  data      TEXT     NOT NULL, last_used INTEGER NOT NULL DEFAULT 0,

  UNIQUE (user_uuid, atype)
);
CREATE TABLE twofactor_duo_ctx (
    state      TEXT    NOT NULL,
    user_email TEXT    NOT NULL,
    nonce      TEXT    NOT NULL,
    exp        INTEGER NOT NULL,

    PRIMARY KEY (state)
);
CREATE TABLE twofactor_incomplete (
  user_uuid   TEXT     NOT NULL REFERENCES users(uuid),
  device_uuid TEXT     NOT NULL,
  device_name TEXT     NOT NULL,
  login_time  DATETIME NOT NULL,
  ip_address  TEXT     NOT NULL, device_type INTEGER NOT NULL DEFAULT 14,

  PRIMARY KEY (user_uuid, device_uuid)
);
CREATE TABLE users (
  uuid                TEXT     NOT NULL PRIMARY KEY,
  created_at          DATETIME NOT NULL,
  updated_at          DATETIME NOT NULL,
  email               TEXT     NOT NULL UNIQUE,
  name                TEXT     NOT NULL,
  password_hash       BLOB     NOT NULL,
  salt                BLOB     NOT NULL,
  password_iterations INTEGER  NOT NULL,
  password_hint       TEXT,
  akey                 TEXT     NOT NULL,
  private_key         TEXT,
  public_key          TEXT,
  totp_secret         TEXT,
  totp_recover        TEXT,
  security_stamp      TEXT     NOT NULL,
  equivalent_domains  TEXT     NOT NULL,
  excluded_globals    TEXT     NOT NULL
, client_kdf_type INTEGER NOT NULL DEFAULT 0, client_kdf_iter INTEGER NOT NULL DEFAULT 100000, verified_at DATETIME DEFAULT NULL, last_verifying_at DATETIME DEFAULT NULL, login_verify_count INTEGER NOT NULL DEFAULT 0, email_new TEXT DEFAULT NULL, email_new_token TEXT DEFAULT NULL, enabled BOOLEAN NOT NULL DEFAULT 1, stamp_exception TEXT DEFAULT NULL, api_key TEXT, avatar_color TEXT, client_kdf_memory INTEGER DEFAULT NULL, client_kdf_parallelism INTEGER DEFAULT NULL, external_id TEXT, key_id TEXT);
INSERT INTO "users" VALUES('d97af672-5183-4c87-bc85-97bae473c4fc','2026-10-28 12:55:13.535748368','2026-10-28 12:55:14.294097501','bwb-integration@example.invalid','BWB integration',X'BE13FBA2831A509B6AE081DC377A641A4600042F136A93DA9CC95C8983E7608E',X'8EB98BABB87FCAE28F05E442165DD289E587665430B933CB3B83CC8C57CE7798A416F7F70E4C9AD2F52A120397B4449127911843B2ECD2E776FF348085A5AFF1',600000,NULL,'2.IWUT4EBWF7UlZ8WKUEEz3w==|HgLuQLtDhNsO+dUOxegtPASMuEMTKNoNF+Rz6H7rwYAv6Sy2/mtO/o2nRxLfNcTYO+1exIYbguQBQ5vZxk95Aa7NG4y2uEf3HfsbCCv9KiE=|TnoUjGfszzaQjI4IEBb/LyXiBlfymru5nHwlhaf90QU=','2.rIjWdrIcF4wWLXLe/FGQAw==|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',NULL,NULL,'5189436a-1092-42b5-bbb5-e85a9c9c847c','[]','[]',0,600000,NULL,NULL,0,NULL,NULL,1,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO "users" VALUES('9bdb25ab-07eb-4882-a1c7-462666e4963b','2026-10-28 12:55:14.168990797','2026-10-28 12:55:14.294132167','bwb-restricted@example.invalid','bwb-restricted@example.invalid',X'E3ACC283C0498263A236467F0B7889F337736902F34FF266C7F6231239236A1B',X'E604609643CE50BA8291135E0C8908572981376ECB901D72C58C8B13C3325AD96755E00A5B6697FEE04EB96E89190205590BF36CC5A0E5CE675D68000B0109F3',600000,NULL,'2.WdRmqcN+7ra8yotw2+xQPw==|er205xFq4i4SB4ikej/YdJoSFLwyJL32TDd49s8jJyCmy1h/wS85h/0sVC7W2e5WDvl6TO7c6mChik2jN9Q3a52fR54LR+9C5n9ezvTN14I=|tTOaqKr82ht3ML0BFkOW6ZW24KxYaeDBkxHpaOwR/bc=','2.p/mokkIXwASexfWUP4ACpQ==|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',NULL,NULL,'a93400c7-a0f8-4e0f-89f8-668de7b4bc1a','[]','[]',0,600000,NULL,NULL,0,NULL,NULL,1,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
CREATE TABLE users_collections (
  user_uuid       TEXT NOT NULL REFERENCES users (uuid),
  collection_uuid TEXT NOT NULL REFERENCES collections (uuid), read_only BOOLEAN NOT NULL DEFAULT 0, hide_passwords BOOLEAN NOT NULL DEFAULT 0, manage BOOLEAN NOT NULL DEFAULT 0,
  PRIMARY KEY (user_uuid, collection_uuid)
);
INSERT INTO "users_collections" VALUES('9bdb25ab-07eb-4882-a1c7-462666e4963b','6a12ea69-7e50-4531-8a05-308947a8e99e',1,1,0);
INSERT INTO "users_collections" VALUES('9bdb25ab-07eb-4882-a1c7-462666e4963b','4f898ee1-1353-4906-92bc-04397f9e8f4a',1,0,0);
INSERT INTO "users_collections" VALUES('9bdb25ab-07eb-4882-a1c7-462666e4963b','93e7e5fb-5d47-4634-8fd6-aaddb9293020',1,1,0);
INSERT INTO "users_collections" VALUES('9bdb25ab-07eb-4882-a1c7-462666e4963b','a14686ec-8600-4abb-a160-72ef4fb01e7f',1,0,0);
CREATE TABLE users_organizations (
  uuid       TEXT    NOT NULL PRIMARY KEY,
  user_uuid  TEXT    NOT NULL REFERENCES users (uuid),
  org_uuid   TEXT    NOT NULL REFERENCES organizations (uuid),

  access_all BOOLEAN NOT NULL,
  akey        TEXT    NOT NULL,
  status     INTEGER NOT NULL,
  atype       INTEGER NOT NULL, reset_password_key TEXT, "external_id" TEXT, invited_by_email TEXT DEFAULT NULL,

  UNIQUE (user_uuid, org_uuid)
);
INSERT INTO "users_organizations" VALUES('d783951f-56f9-4dfe-8460-e9233a708bf7','d97af672-5183-4c87-bc85-97bae473c4fc','5f25abf9-a963-4086-81b4-48e48ba9927f',1,'4.WFUBWyNnZdUSn4AyE8JhvsvW/vzwKUqJ+GxF1pzP0bA7uVzqp5kykJl7jj6/TKc16xqpN+S45V4n9KnGCR1cKr0ZModDUC1BH/73YI7FRXUUcyxjo/Z/+FAQ3BGlfaVHK2Ni3PSxX2Z2jFYyhoYnLrjKJvQdThIelg5YLRdluPU0yCXFrBwyp7SVczG1VC59Pog5IHgOpCRUMyWqtu/Qedrcpsd7y23mzowDzvMlBltloaUJHSqJ8p2l75/yZEhE65jcEamOTdGh7Gre8tlgO0sEJ35A6MlywIVFkW35qw1jul8oqef6W7+y3KFamy8xNJgnQQcc75q85W/yBXQGEw==',2,0,NULL,NULL,NULL);
INSERT INTO "users_organizations" VALUES('1a4b2779-119b-489f-a0bb-217fc51b0edd','9bdb25ab-07eb-4882-a1c7-462666e4963b','5f25abf9-a963-4086-81b4-48e48ba9927f',0,'4.l3JxhXkhyau/PIg6OY39xZ4TQYkEAtMV55tBjELCbYbEJb0rxihLWRCx/ECmrQbP+RJtzZXur0Y9HziYGmSbVQW6gC5lzm1R7Aqaki/xJZj37QaCVO3I34JDBl/+eMeFzj+GR95pNj+iJz7mvLv5jv9zQz3kk8v8YJcAfucq6wBcNWBldXykatFB0AmuQYfILGXLa15h5YZKowvySrpbd3YA7P/9DhtJE5GYPEWbRMSARHYQg5KAxt4Pg8CDWI8eXKZpr3FxTqjpBKhl9vrlVkQdBtSJPARB7b5Fa2Sz5U3nF3exDuKGMgLga+STZMIOl1pBvRbAgSBKeaigzWTr1Q==',2,2,NULL,NULL,'bwb-integration@example.invalid');
INSERT INTO "users_organizations" VALUES('3d9ea39d-6e07-40ab-9486-2fc6365e9eaa','d97af672-5183-4c87-bc85-97bae473c4fc','e6e8aadc-2d28-4fc5-b1da-d02b97ab2f17',1,'4.j+6oPUpg1xJN+VJoF26BcfEzqJeaAjI8O4JaakUgIBItF3eg75J3KXU75X34nr1ryLGndABGuTUAMedur43FYGDyBaknWjCLNXr/kvsXFdZdhAqhblNpdZ6x0a39atqM2/ZkrMRliwEc6ofm1IjRUaitAuJ4tZnR9+QahPwmnE89f6fJYo326+HGjsEeEK0LCuefe6mktp3sEy9A/fECJtxv9dVr5RhnGZs54qwxNPqARAe4wxrsEpN+dwlL7tMEWu6/E49tM90Gu5yJDCSe3bTdeQqPPRyP8sYNS9Qq1w/6uKihZlwAfmhxIlaXSeMWnqWGGPCkh6MzBnkCXNR0Pw==',2,0,NULL,NULL,NULL);
INSERT INTO "users_organizations" VALUES('52919464-c2da-4a23-98c7-b6955b2c4eec','9bdb25ab-07eb-4882-a1c7-462666e4963b','e6e8aadc-2d28-4fc5-b1da-d02b97ab2f17',0,'4.j/fzxlwjIprJsc0Bjc0cdJ2SToBVgaIbaAlshmY1vhDjAnTVkn5UFQm/E1pbL1snUstalBQ//Tq7CU4YR2JVp0/+lIYOH0GK4Oo4T5nEbCgmQgKRoWAQ3/Phs50aamYztH86FaTirOGPhJbUtW7yin9W0XbqzMOgIL8WpNA2dDUody6xe7yL6skLbLyEJMZGkKhhGigto5RWo5iEeilBjGAJ58103ZTMtJNW2Hxn9rQ3FblCrfMWRs76MV5mJiGD5Vtc8Vvq+XTQ4ms58nvqjwqsYJT7KQV3ASNWemRV3NrOnqEYBGhe91Wuz3fyNenfXcgSbUEW2Hz5KuGp8SQtpw==',2,2,NULL,NULL,'bwb-integration@example.invalid');
COMMIT;
