BEGIN TRANSACTION;
CREATE TABLE __diesel_schema_migrations (
       version VARCHAR(50) PRIMARY KEY NOT NULL,
       run_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO "__diesel_schema_migrations" VALUES('20180114171611','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180217205753','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180427155151','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180508161616','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180525232323','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180601112529','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180711181453','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180827172114','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180910111213','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20180919144557','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20181127152651','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20190526216651','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20191010083032','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20191117011009','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20200313205045','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20200409235005','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20200701214531','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20200802025025','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20201130224000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20201209173101','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20210311190243','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20210315163412','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20210430233251','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20210511205202','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20210701203140','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20210830193501','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20211024164321','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20220117234911','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20220302210038','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20220727110000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20221018170602','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230106151600','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230111205851','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230131222222','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230218125735','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230602200424','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230617200424','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230628133700','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230901170620','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230902212336','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230910133000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20230914133000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20231021221242','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20240112210182','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20240214140000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20240214170000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20240226170000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20240306170000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20240313','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20240605131359','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20240904091351','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20250109172300','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20250820120000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20260309005927','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20260425120000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20260505120000','2026-10-29 10:43:08');
INSERT INTO "__diesel_schema_migrations" VALUES('20260902120000','2026-10-29 10:43:08');
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
INSERT INTO "ciphers" VALUES('da2d518e-46eb-4391-a4ff-97c965c3c7b3','2026-10-29 10:43:09.347608246','2026-10-29 10:43:09.347710662','e08aec66-84b3-4d66-96c0-794442636726',NULL,1,'2.a1GEq70mKJRzGqmeLR8A6Q==|Z94WTVYDR7fbE1P9L8tPgA==|MtQ9gVtXv3FlqraKKj+NU8EEpO1Gt7ulGCkwooTW2Ls=',NULL,'[]','{"password":"2.y6TZcznYHqtIRjt6rONu4g==|ay2QOLtydTFbb5sGTm2h8LMeil3zz1Ihr5KEgonKA08=|8SxEG5FJxtdoxXunoFGfd9N+5dPjV8prPUOMgrFiDDc=","totp":null,"uris":[{"match":null,"uri":"2.tu94oazIHRDDKH65SjuIHw==|u3lHtAcnV13jOxRt5rRv9retSL1HNidmNVVXIzYw8J8=|fejYo0rvKTTHpl6clZtRg17GTHs7TMdEtRyYqd8dORY="}],"username":"2.lgc2njWT6zV1KuLqR7IDzw==|kaVCpzcZbmc0PIyruQZkgH/2g2APJfCuQ5g1byZpU7w=|jD2jZyHJVaMGKrirKmVR47EKR5IU4bRXWN0Pj5/Rs2U="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('56e875a2-5a31-4a2b-b4c4-6dfa656a8cd3','2026-10-29 10:43:09.348935911','2026-10-29 10:43:09.349000786','e08aec66-84b3-4d66-96c0-794442636726',NULL,1,'2.hU/3R2zWd5uXRH5/Fk6VKA==|eVoqcjzt9poi84sK/U89wA==|6UTBmFyA2uPd3L8lC658xMcsdDc81LsmHAq0ly7tVNg=',NULL,'[{"name":"2.VYY7kk1h+CnpT5MyI6njGA==|cczyJZxKSBYKyfwMfDIhlA==|qjAgaLhs2mnKNCnbn94RBUz2GiVHlXwnKqpoPjyS/xo=","type":0,"value":"2.fkq9RBVTTXAMO+90Jh0bQg==|8W4iPK0qUEPRxJhy33VS9g==|z6bcRz02ke8P7RyT8IZKOmjlFmzTAkIPTTcoJnwjnZE="},{"name":"2.bmIKU96Mro6Jp1knkJ+OqA==|hR8Nipe4OT37IJVJ/ZtILQ==|z1wN/817YVkTGDSe9KQ+LSR9gbv8AUeqLQnJ9H8R1IY=","type":1,"value":"2.14YNI/hC7AQ+hQiQYeuv/Q==|+4hwiCMa/bpXmpsJ7NFMBg==|JBaIhCZGctdSNAg5NgrOkoMHkVjku6M6d25mQ0aRAvE="}]','{"password":"2.KjCLRxvs0QcwqXTwnT8wYw==|BNqf/OpClUb+m/a2cHWlGQ==|AKvfszRoh1qJRcyipkYZLBIFL+rhru2eY2cRhPqdbEU=","totp":null,"uris":[],"username":"2.eqhwzewAZCALDVypg0C7Nw==|c3U5y7PkwcvlSNd1qBIqFg==|Ec6Va9/KZO4H1A13+dOTdfQdlKht3yW2tmWn7k43pE0="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('5acaadb0-a930-4442-bb3d-f07f5cd5ddfa','2026-10-29 10:43:09.350090784','2026-10-29 10:43:09.350143242','e08aec66-84b3-4d66-96c0-794442636726',NULL,1,'2.DXp8POgLMIaWvvHrAQGugA==|ceRqYqowbjNbc3fC0N58zA==|pKEKGm6XBWxJjg0hILBdGcFBmx/aQCSYeyf4NsHwyx8=',NULL,'[{"name":"2.F3P6Ok23CC5bVwG3IkENkA==|ZABUz2ZnJdXBlM4LqV22rg==|J2XgIGs4D1Ovw0bI5L/PURn6gcJE2IvO8FQIYLHTbfg=","type":0,"value":"2.lTl6Ez9rdgfZ1+11m4XQ+A==|eXqjKqiR0uhE2UMTa2MI4w==|sbZ4vUFfBBOmvUxfx5n+gvSt7f0qAUIZ03SKLr+bTnQ="},{"name":"2.IKvnMyFaIK5ilfXdL7/y8g==|MjnlBkw5EtjAi8Pbb1JLuA==|SBkIuDzT9IQxKO0lEUxgDZQeDkc/vakXRx7cqBDkIro=","type":1,"value":"2.UGSKYfGlvg4FIoNGA0PuCQ==|ldXWGV6IF5eK03ZKjLpKPw==|DLZKo82aDk2YoDH3JjlI0i4JnsjdDuXvQGXbw2EiVqw="}]','{"password":"2.oSdRJCcjmfPHxzqZQzRAqw==|ERGuc5exOn2Lxdtl+orjjQ==|Eydxmz7mv42VmhY+4VQAZjin9KGND4o0dzMmt3E2ze0=","totp":null,"uris":[],"username":"2.qWt5544XFivcRCDXrasAtA==|xE4L/iIsGctaqCeFeBceMw==|Aiv/ewLrkVQxAQNFH1+br4MSOswofuK8cTNtg7ZcCY4="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('68d7d179-685c-4912-98f5-1e90a98ea70f','2026-10-29 10:43:09.351159991','2026-10-29 10:43:09.351234074','e08aec66-84b3-4d66-96c0-794442636726',NULL,1,'2.BEYyZl/LpapuKslJ6pNktg==|T+EqzRlLdjJV7n9Ua4k+OzeR16mSMsBK7ONCfM5ZQdE=|av9QQ90y6DnsDK45EsGss1GF8Ta7aB5Xk2xmqVl6AnA=','2.YiWEHxKZjLBpHQmBi2q9Uw==|MmDo/o9iIPPGzRnpjFeJCT/g5o1JhbZn7JL+HmAA5IU=|4SMlCwtfc/FHrUD00SxOa5J9KbuDKUSeNPbajT0Whpk=','[]','{"password":"2.7sz4BeLcERV3Jv+X18PLNA==|MstKep6axG8tIwXXO9imLFkvmXrCdrfC8TKIeGSe75Q=|Abmt92+iWJtJZWfoVM7xfpUxjTdF+euIegagZUbBCM0=","totp":null,"uris":[{"match":null,"uri":"2.rgwGcBEQ5tXWIjf9QrFmKA==|dERNGto6bSSipmsDn49OyzVeoMIL4jH+xUJBj+jVv1Y=|JPwCkuTOD7+X/IAP+wmjhPDiftb9tow+Qr7N+A2fp9A="}],"username":"2.FGSkNqLS47hIHrsyCDI7OQ==|f34YDxVZe95kPrWuFH2DrV/ulmaHodxqiC5ykBO5UsA=|QY82RpJ2vg1aI8j8dYnkuKbADv4EPTSl05jnd8h3c74="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('836aecbf-24ef-4b55-a3e8-0d830e8c5923','2026-10-29 10:43:09.352311697','2026-10-29 10:43:09.352356197','e08aec66-84b3-4d66-96c0-794442636726',NULL,1,'2.zRAVmyK2rhYYNX5pmlSezA==|V4ID6cPcbqBHBenR4xyg0g==|tLGJBHV6t9zssr6cIKvf4aZJWJqWK0E6NalAVvHiwHM=',NULL,'[]','{"password":"2.YbFmsuwcuNvCkkEXeuy9JQ==|VqR2pHZ1OAG1AAak7xRSLg==|8v/XWIMsiBKu7g7ou2PIrIw0cC/CSHr1+g0abKHpEgY=","totp":null,"uris":[],"username":"2.EGsv2lj1/Tqf1628hLKx4A==|i0KQy93vPLR0drTCe8rCtg==|92qGsD/XmEZtPs7Nwcz37FQwuBTSSI3FmABKcsrYhyc="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('8002442c-cba6-4363-a28e-6dc40902e0fe','2026-10-29 10:43:09.353236571','2026-10-29 10:43:09.353309738','e08aec66-84b3-4d66-96c0-794442636726',NULL,2,'2.zlgpCdPT+t35SnM8wKlOmg==|vEpANBOWqrBmWJM3AtE0WN0iaCpWrCI1ztyt3Mqiplg=|+tKOSPjWFhdzJ2EVeXivZqWUcX52ifUa/gl3RiE2Lsc=','2.2sUNq6j76YxT0VvxyW54uw==|3BK1Ym6spTgvLvYjr0T87HH7rvrfUkWA1xOO+mw8B1Y=|Z4rzD53QSM5zgMmS3qEcmg+XONXEfRwig0wgLSEkU1c=','[]','{"type":0}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('9b26b892-31d2-4604-9ace-05193dad3255','2026-10-29 10:43:09.354356986','2026-10-29 10:43:09.354413611','e08aec66-84b3-4d66-96c0-794442636726',NULL,3,'2.JbvDM3x6uJa1wSa8ZGe4Qw==|tpaReDsMbGohmnoYXCxzW0hzGPafxrLFNkicDUDU3Ic=|5wuUaC+b9rMwEpVM0A/DdFtByedbO9ydJsBmkn2MqP4=',NULL,'[]','{"brand":"2.U5SjKqL86f58ZfmphAZLgQ==|+gkCC3pweeRqEunljrCfyw==|LCDm9b/qrh7TzfyojXayMi/IRqs9yMa62+AOf+R/4T8=","cardholderName":"2.kvp6CeodWvnb5Ngo81CK/g==|qGIeKyPHTRdrr7xukuMpOQ==|zNqfx1ATUdhrOi4a/s8S+QGiKlR3rrSr7wsLkrThfhY=","code":"2.zkk1JeINnzYWJkT4roobDA==|9I4YLTgGJ7gozCA378TjEA==|N8nSbH31wO1Mxe2C8aj0+iSpgeTsQVDn5Rj4qwrEJxU=","expMonth":"2.YNUyWDlpFDoZaQ8CevxV6w==|UWjSkfNI7o7nhIcRkRCTVA==|HMvyKSrbjzexTum8O8P8qY1Nwr/VX7Nv60rGQl2ohx0=","expYear":"2.dDuQ7I4I8iHpKqHXnwURtg==|5llsbRZ7DzGHX0wv1CdXoA==|E2HG2nom2ggs5Vjsah8S4gCCXeqE+9NbRPv+o2hQVsY=","number":"2.JeorV+hmSgxRdag3YB7crA==|QFWPfg7PLRjw7xnPOjQkF+ea4ttziQFswX/r3hrUTeY=|frOqCEhS7Nu/DYSl1yedOF8DmBkVMaKz+RpCbp5R7og="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('e993b47e-3ed2-49c9-9f45-cde5a757e39d','2026-10-29 10:43:09.355438485','2026-10-29 10:43:09.355513193','e08aec66-84b3-4d66-96c0-794442636726',NULL,4,'2.5otzcZNVAbAO4VPVfzUDiw==|e1Bb065ilNMZlGNr0qXRXg==|/Ja/3yl6Hi0fj5CXP87yMfbmmmgKmbQQ5b9MtVSnMFc=',NULL,'[]','{"address1":"2.Fbf1BXZjbvIe2I8IGYxDBQ==|lcHIQ2H79Wd+4Vr1cn/YCP2dKF3g1B8A0ioXSBfz+to=|ocnhCv+XeCiin32vgJ5e8KyhN7OkVVALBuOob26Ib28=","address2":null,"address3":null,"city":"2.FY7czqGpC7kz44Eh1ZcJVg==|3qHJSe/tSMu4X+Cuz+0ylA==|elGmQu/2GhKP+aj6zIFszy0Z6DZZ4aOuOzMWFJBL59o=","company":"2.BZ5I6czXh2s2GZXY5SRCVg==|0bSPusai+aIaw+abt6hQYg==|oKlY10vI+c5LSkccfj+FW7OCdBWpr/nCZdcu2746qn0=","country":"2.XDj5Rw6CuActwjpNhQD/Og==|oxefv+lQIxckhI/JYEb8dA==|E9nVImFUQFmdlICzhseIjas7dBEyf6nhoCbIyJ09Zcw=","email":"2.LVK6qXqqnGBsbYbITge0tA==|5VK2t+bP68CqT6Z8KCof8XYoAlCrIjwc0nzzqTBLS9g=|0I4Prt1BztP6pC7xzw7jS5jqVKCZYXkJ21u7+R94Oas=","firstName":"2.GgQ0Dwnfy7pNJmz8rim5ZQ==|T4P4M6Byxi5gu0BQF6D5uQ==|cmtzDkIz0Fc9IGb2mTu2VaCVT01NXDZQaMvMhF6HLFs=","lastName":"2.y74blo0y0p/wLCkojOp9Bw==|pYXp0Ek4l7FRdD5ZveE8bw==|e0vZe1bKGgILzlZFg6e8TEruiFC/Zvr4quaC0XJLJWU=","licenseNumber":null,"middleName":null,"passportNumber":null,"phone":null,"postalCode":"2.EWFYO4OVzLxKs4o3jbZDkA==|YClD1kJZ9EN2ppGV6H5rDA==|AcGuvjnw4lA77/bDqF2amu01apmRl2jCwCyFVO2jAuA=","ssn":null,"state":null,"title":null,"username":"2.nq6mhIA2PmPAWE28/ohxnw==|iGgRa5wlAqwW7Vd7qBGx+w==|4ZPExJDVKRW8gGVxKUEIpccQMXRx9mQjrmbTmocTe38="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('40016660-713f-470e-b21d-ee9b6a032129','2026-10-29 10:43:09.356569275','2026-10-29 10:43:09.356649899','e08aec66-84b3-4d66-96c0-794442636726',NULL,1,'2.8QAUpmfEMTcUH+pPpzkW9g==|uFp5Ly/gjeeCUHGAgYC/Z+zWIkpywC5e3vp7L77dq9Q=|Bk4WAbLpSmHXo1rQO0UVXmWSFmNe82cHmJlJIIVHSbA=',NULL,'[]','{"password":"2.wLnMyVM0zxZqHgChKrUd+g==|62gyGScKns5N8qUh7hPKXDoO59K48Gvp1+sOOoQf4Ho=|rWcXZZepb2zVKad8BxjnZEogEtXZ9/a2pZEwpfttSDo=","totp":"2.igoH3ieVRcZVVYLJQRxCDw==|MUQ44uaLFv3mZAuD19ohbrbyON3IBG06ilyvbpRs2sP1lp82lOcCvml7/E1S7IuI|15thtHiP620aeoSuJrpZdaz5HGJM/qCgCgm7FrbW18A=","uris":[],"username":"2.fxJH8Zw10/Iuw5ZDxfvB7A==|Lc3h++9dMa/1cnq7TFsYYOZASIUZu4Zo+ypHshks3n8=|p7dGxD1v/AzeEYCW+XoVi2ruNGRU8NSVW8Y+523M8ZQ="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('aa1940e6-2051-4235-92f5-afec0d73ca19','2026-10-29 10:43:09.357584440','2026-10-29 10:43:09.357622356','e08aec66-84b3-4d66-96c0-794442636726',NULL,5,'2.7baKNCWPEJTWjBrAWczwsQ==|hMLLG0x13CSOZAc82wtOitU9aODJPj7m+PfGRzN7miM=|qG13GVUHR00bea3G9wPiPc5lKOQO0715RDKEYldrNfE=',NULL,'[]','{"keyFingerprint":"2.FnyFUMgbAJkf4DFqP1mlww==|jZM+wtuY83els5FhZJWmvsrTd4i1vP5RXnMrrwJJ5vM0RvS5wz4LkWvNNJef8Zxjbbq42DHEMV5mAkN5oPV3jQ==|4VKH9aqkn53vJXweAYWlc6rXlzqfhBhjlhzajwyqZso=","privateKey":"2.7otKkqfIIMG2B9pZBXQUlA==|9tskppVh4q8814KCR1fybKNZtZJc5soMqdejTKuFqWtHluBJR4icmW2GKxR2hyfEe2VTgcL3DNdXldRdH383U0XkWO7pf04rD57RsDwv3znrzJMSofd2cD2mYaWLFbX9WXM8bhNPYwV+Rp8okTTUQ11n1hTNryo5jtxGi3no8qd+q6CcVPGc2s063HSGSjNAJKyPOEMkXagTTq62zl4STk8B3lBUrxY6EwpiVrhbm0EE1fM85SJmRQYuEEi/TBB9BpC1DxYJm4v5ninQzxJ05IXKYzspBRvLYS1tlCbHsep+lCW+sa+ArGf4Sha+DoP/Wbn2BUuc4lli27HGFw4rQFjXwys1h7P2sVhHvQDgt14kKUaCiLFHu0fAPMgd2toGezS6JHYHkUo5VVqx767LeJ9ymzJwNaxCMWes027eRtJRwaVH4KFzFZNoh2xuFeZAxnIg6wSols5tZlsbdbm2U5fLFw873odX1KD+NgJytuqCl9EwjWakJ3Wr6/ad+xh9VuCqLnoPCnjVULi9Umns5Q==|cj2pEZKoRHy1wXCxDwtJHb6bVkgdzhGnixIv2uqhGL4=","publicKey":"2.AcOFh+pS1vZHF81KhmqBgQ==|6j/5azot2OO0oEBmHCQEkWP5/r7RiAFQD3axE7JASh80ccq/q3vooihf4gDj5FGSgIUPXuB99Qiwgomvb/Y5lxZFAkxyVh66HDB8KVk0LDwsGqAS39t6qCIFBAPgNSt4|vSk59OcFBk8WrGHcHzwdLQjMOBzFt26Mmh4A1Bs3ukc="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('24a3a7f9-7ec2-4b3c-8a4b-4b3a190ccf6c','2026-10-29 10:43:09.358618897','2026-10-29 10:43:09.358663063','e08aec66-84b3-4d66-96c0-794442636726',NULL,6,'2.dqmqT4jGdMVve7QF9wQikw==|X+SodvZS2tEVw79SmP2DeyEK7kTaBK7HFkFWlbuoTWo=|shymBm4/vuAwcE9tFgzq7zE+VWHH2NN7GkZpT6s2flI=',NULL,'[]','{"accountNumber":"2.CYjLVOZR7PNPmiPTFUG+NQ==|qCRtfIHsG2FVmZC1qJpb+g==|bkoFGSZqVPF2gk77kkeMjqB5faOVSNZWoOptQc8TbiU=","accountType":"2.ceJdyHvEuqpSirQtvUmn2Q==|cJ1I/ZsXJOm/DR38QghTdg==|5e47VT7PX9OwK4bDUsmaU3OhYvi9pZMyokoYhoneDgU=","bankContactPhone":"2.UPjBpUTAtRCoFfhbx95EJA==|EnCsYLfCSK/LxcNz+CPtQA==|kowws9j0de2O+XrGHnLPXzihQMtrRFDqQ9MUmF/HKsw=","bankName":"2.3Zo9fZfbTfahLhsBbSuu8g==|qooXV0L1JhSVgFv+u1ILvw==|MFHwnwNrBJ93BkWr2CvucVUN3Y/LfGNmbTYBKAS7UN8=","branchNumber":"2.JV6yXOzSFv7+r6zJ3ITmiw==|WlpfajeyjXSO08KJIVfPmg==|bMuyvo3dyGb+R+crxq7ZttaaqQPQlvzKH7RS4aEkf9s=","iban":"2.FbPsDRKjtpcTEA7K/9Qydw==|44eKuANFmZCwlZqmU+kdVw==|/CUeOnXhEIsyawG98eI/wrIUUVbBXvB2mCtzDvYIgc8=","nameOnAccount":"2.hZ8ZOpoERkgrko0NOfMN7Q==|6bAcfFRx8LgKCYtW3ATneQ==|1Dz820CL27w2yjT+7bO2pa+BmVYru1Ca82Iwkg0F374=","pin":"2.p+2Z4Tt+i26iHSS4Wgo+Fg==|sV3pStteKoCLeMHylYLGpw==|7MpVtvN+LbOOqYH3RNVcrcsyggDYB2F1Uw1W8A+qOYo=","routingNumber":"2.WPlWu+rHLJK32ZLonF1bvA==|Kt4d4bmBFx4h0wScbclkoQ==|aVEYE7d+AqRyf6/QD5UXLiOevsplCqRAlqStIWE7C6s=","swiftCode":"2.nrlEKhiV0sg+w+C5uBBQjw==|I8ym1DglKFdA2sWvej4fCA==|uPYvVpGPIIGJZtbUFdyUsmZegaJALHow6Nj1b60RV78="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('787457d6-8bd7-432c-8db4-dccd574fdbb9','2026-10-29 10:43:09.359627687','2026-10-29 10:43:09.359692770','e08aec66-84b3-4d66-96c0-794442636726',NULL,7,'2.isLaCXq5KCFx2Kuexhwz9w==|wcVcjJEnEG7XPhSbGBZkxVsU/skH7g2hlx1cANa7S4I=|9UiesbfnP8iDcpZAkWUq0WnrdA3rto3V6YdQ8F4VGTQ=',NULL,'[]','{"dateOfBirth":"2.B+PC2EJw2EdC875N/RtapA==|FSrbKDA00dhzYCx14BfLbg==|h4v7lvVZdPBZubIGBBd2EroXstolj2w0NBzkn9k6kIg=","expirationDate":"2.cSvQSYuzPX42yDQ1b/BVnA==|dYYX5gmyHCsYsXPYgGVd3Q==|wFssSVK9vHXTsZicJvmCQtVVhjidqGfUBsNetI3TUWo=","firstName":"2.2JsZ8PG86S9J0cBpnMKs2A==|vFJ8vLBOhh3JrcYfAUDEUw==|SuNGFZlHLlQL4WengfesP76evbjTcnCp2/Od6Sso6mY=","issueDate":"2.63+tUVi95OYoXGz1Xb/RhA==|QYlozIOOkq0EoGOtnHh+5w==|aDlPjGdOyECMAA5iNCEg15wFDGS/gn/SZdReCmtCenE=","issuingAuthority":"2.kalQN4A21+/GMOydNf7x+Q==|O6rrJ6vcAQ1hs8brLER3vA==|QwYbmQCD+OEU6No7J2PhZQnDFQSfiOEKXDPHSjVegho=","issuingCountry":"2.ctJntwh6QPvFJ1uuWMs3Qg==|zM7YSZ1C/iBy8H7aSqJf+A==|ik8Un5aw27TvZLq7SiFB9n/uVbUrfLmZFapppvAkwkE=","issuingState":"2.Al0qTwmab4LSiQn3sJ2+eA==|smwNYseIJTvKj1l2jXfnwA==|/EkqutApz9VCINOTqCxKb7PQpQ7Jr7fbD9kE/wSs9kE=","lastName":"2.2knVszNp6xNGa8Ngd7mZpQ==|y1yX6HOr3NyCzBiju9R8Nw==|StMDVUz/x2KDnP7r7jiTHI3C7jHMSvhBMSf6iEsIYAc=","licenseClass":"2.w1LKjf12qYQPYvgzjDzYNQ==|TGHFGf11gKbbwms6recwdA==|mrBHihnMkKUVaL/z7/wOyHFjZQKmZQUCefDZ8fnBqR4=","licenseNumber":"2.4hKgKsVaRJShxE22gUAPXA==|W1ud4nmrh5XLilN/d5N1eg==|IbqK7GAEw8/UdcNUdbCmjpzcutX11txcmL3otpfqSSg=","middleName":"2.rpI3FCIjveUQkoOiVyE/EQ==|yn+lvzB8MUjN6K1m+/v26Q==|mhTYUHuJNKY2zOFZ5kNSv6l93gss7oWOiKG7SEL6lHE="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('01b4d401-7f30-4004-b4e8-ffd271bb1e14','2026-10-29 10:43:09.360803143','2026-10-29 10:43:09.360854560','e08aec66-84b3-4d66-96c0-794442636726',NULL,8,'2.iABSE1w1qwfxdpZnhBzpOQ==|P7vR1pZ8EyZ1iSYtjGskWQ==|HaA4W9Zujc0SeQBH2NxQqZz/z7JyLfq6W+BHWHYg4FU=',NULL,'[]','{"birthPlace":"2.TsZlvrbjGLcp5kVznRhgZQ==|f/gDBv50aoWewiNYCJRWCw==|ie2u3foJW2W4SkSf3sWW7t3UojiJIdnUlv5FUzoD768=","dateOfBirth":"2.j4URiS6iQ/yqi/EakwUFHw==|zLsfyE4v8kQRBrMsPi/V2A==|bRpFnzi+z2wgNgRDJ3Q3ebg3vW61RTEDHOX6+6K36e0=","expirationDate":"2.fGOBv7Pb6VlFsgapC5q6Ug==|3M5KnZFSYOuUF9+p3U+ydQ==|yEPNQN59+i7giwWbiKgwSNmu5Bcn5ZbODLYKnEz+780=","givenName":"2.jqPKQWj/U2H2jXlxhghW8g==|G43jOEv0tbIf5ltAOU6vtw==|AzRWmKiXylIMnSjNUOdfNaQBg/RSi3DgerhI6fAU21w=","issueDate":"2.4WhQpqqA9GirYiZY6j9t7w==|xUnHNiV65m19bk0mrEqlRw==|GKBK6v8z9n2dSHUor+MDO5p5YTpOE1oncJswQ+H8r/M=","issuingAuthority":"2.q7zjmMTE8ds0PnHFhictiw==|w8QXo/WFIwH+f7kidFsXng==|XQWilM8Jdzx8C8WVSyO7IE3Ig1Zc5pZBVLFtNz6Sf5o=","issuingCountry":"2.6BH3Zn7oPbIWoYGJ+vTcdQ==|9/ffp4s7X/dXE554R28aRg==|euCXjZkSuFngXnL9EEqCablufzo6Ocb7lV8QuFwWgtw=","nationalIdentificationNumber":"2.spwejIeJTCDpteyN0uhDdg==|JudvofOAhCaEofafEI5fLA==|Ys7lU/hTEe3kGQVcb2005UZkBGrkseLMO0soIZ8Be+g=","nationality":"2.XGesVQAu/Obs+APS/IUBXA==|T1L2QhHJCtujp9gXF3WNKQ==|/ExdaVtOfSiyFxCD6KJTn734tNg2j/m2RhcE5t+stgo=","passportNumber":"2.ZrdSSAkjwgXIHjzYoA7fDw==|nFt1Npq/4BhLLBPtf8OuhQ==|wogM70y2dW4msNYL5LOTmxG4COF0zyDYVVbTaAXBvu0=","passportType":"2.CywJYmEzZ6OL6J1CxZIAfw==|31x5tiIuEnj4EiIBGWeDFA==|e1Kh68oi/gor2EbgjTw+Shjnv9KbIguRFYoiHSulRH8=","sex":"2.KONBj3D9SDZT2Hq08m1qlA==|4v8dn6aHkLfm9gr+s8H2gw==|kndeEV+893W5kWHcncdPuJtFgOhj9XerNJETu7bGgw0=","surname":"2.x0dbqi1H/Aue4ECsQQSdpQ==|E6QJT0xcfsstR3flgW4Jaw==|8ZE/7nHBNcvJSO1CkH3MZf29nq/H3U+3l/2Ebwu3I6E="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('32a99461-4243-4fe4-afce-ee73c2bec436','2026-10-29 10:43:09.361801642','2026-10-29 10:43:09.361852225','e08aec66-84b3-4d66-96c0-794442636726',NULL,1,'2.GJNhu58+R7+HVpdFInrb4g==|oj/kYHlBRpQU5d1kZNRWYpS/2lRXBkwoAnMIgiHWKHk=|nRVz2gyTBmblzZlTNSU9i3mFM2jiRFwu69irzMNzqEk=','2.9Fq9qGgVA13/hb6wkJbngw==|9FBLD4qJ7XrXF+p4YCUF0leJAKPUSgv0zeHZtbCUVNE=|S36hPUYtYkcPkyWrhguec2f6+IIpcKYTxtbSZJyE/h4=','[]','{"password":"2.DqH9T1B0ldPcCNcVfRQOow==|3Jmz/Z0ZKb6pgPY51IBhRclDw5qMZGFNzuME8PGyWU4=|Z5zEIJu8NsYrGHaJJOZdcOrGM9RcHH+d39JCK7SzbMM=","totp":"2.6xpZtCsiBTfqj2UDaApJ1w==|1SjOpXDPmrnBwWGU8t7ydyRgh0tSp3lpJsLOcp86c0k=|67d8nymvd2T4ZeEFxJcHXqYGlfvKOi3g0F/c5nM4u4g=","uris":[{"match":null,"uri":"2.d+iPF9KTk+TiLfVRbS1oUA==|560BkSKgWd2U11IVmXBHDXAzcmSr9RQUAvpq6DAysoQ=|heMqZ/EnMkvGL+Gwy2zdyK2d7rxJjyHPUYmMA2/2wUE="}],"username":"2.Wy7lQtqVIRnXgw2n1clp1Q==|iOiVW/kT273al5hfzQnr6NyyQnCl+bT9bb1fdYvvkb4=|nnUrCklUp28PMF7JoLAfjAD9a44OIKMMqaWbIpSCDb4="}',NULL,NULL,1,NULL);
INSERT INTO "ciphers" VALUES('5276181e-1607-40f1-be3a-40836619ae22','2026-10-29 10:43:09.362850807','2026-10-29 10:43:09.362912640','e08aec66-84b3-4d66-96c0-794442636726',NULL,1,'2.ZcMfrJ8KmKxm7Fnr3RPoZA==|9J4YMvOu9fjdkznAxw9neTqgpR/gS2Oq42jAe8UX1T/OwKjKpguOMM4KIN50H+z5|ZKnd85wlxYq56xrUIEZgIDvsdrX85NBnAmab+gk0Q6E=','2.vVDkvxVGKZa31BXoYxedOA==|ERjJuP8GOj2sU3dVO07TbKqugRiK6cOJXVYmsnXVdPc=|c0lmcuhrP/MQQxS+qweY77O4Zy3npENL51jqJVm+I/s=','[]','{"password":"2.pZAtzNJDpz2mW5oJDBcyoQ==|1OTmWhIj+f/hTS9tSHjt7l2PBwt3nhWUFldcWswSj+I=|N8zs4alg/LIELqMiuH81LPsDfYnKbYJ/+gFTNb1T23I=","totp":"2.CIEBcWR7iKX9dSlF1fRU9w==|4sSn/pEehL6fhOXuEIpXnDO7BVquIb8foUFEnC9ex+0=|7ld5LrGnzGexgGa7JfHUVQbyR9ieKmvSPW1wPECpGF0=","uris":[{"match":null,"uri":"2.0Dx32bERK6r8b2KvgTs+SQ==|ZXcYF3zSBh39eViVoDcUrdoY8whMHv+ZRaMKzvAXoTw=|5wltrCGW1y45l/+5deAjhcYBvDy1DP4/kIl9+9OZjWI="}],"username":"2.U9gIFiczhbK1UF5pjfMktw==|qDJ1aJWbsCxTa9y6pmnNL712UhW+frMfoH5iiO6Ckr0=|G0H9f+c6StsuYO+tCdc0LyNmwLKj2Knyefttys8EMqM="}',NULL,NULL,1,NULL);
INSERT INTO "ciphers" VALUES('040a4360-e1e7-4160-9924-f0b1af25021e','2026-10-29 10:43:09.604699166','2026-10-29 10:43:09.605203582',NULL,'b98145ab-1497-4ddf-bdcd-9b12c4c141eb',1,'2.9weEZb4VokVR+mkEVh3nzg==|bKbjoGi2YpJUXM0aLQ8/DmjqbiD6zltuEBSAbYUsNKI=|n/mP6sdyRCIrLcvVcmdZIVSq7RzelqmJBLOf5iiFamI=','2.1dDBOdpimP2ubavecmg5+Q==|/FTCpzxL5TggMMVFKDGAgKj3H6y4tuB7p7W3MvVtUIo=|PzEcW0aF19cu/qyl8SuyMI7buXd5YBt2IZh1c1503n4=','[]','{"password":"2.e4hKnkKMXiKjC+JUo2TbPw==|SA/VeUl0oNjkYWa2ZRUNrw==|qVJIXJ2sSw2Y9s+yV/M6xMhU+PhycB29fyD9YsFnWSY=","username":"2.L0yRfhBO2u8eQeb05XymiQ==|8WMpOKeIfwf+0LdQNbZ3OQ==|7IfVwlRwdJDBg3eUU4fZcERQ/B7Dvyw2orFLxBTGVnM="}',NULL,NULL,0,'2.VTCzHcVKcG/MUdgescbuTA==|YFKk6m+OdAIAc5LiFgJOska5zTeSUQpNBLMvd0boT8asyYZp66NBmt8MDQA6TOyPdQsTGRnHI9IZbuzA/Kh/5ldcD+CLjY6osj+DAL8HLmY=|9hsLa7eE7W06EKn79x0A+SCn6nhcFQ6yT1eG4T/K1Vg=');
INSERT INTO "ciphers" VALUES('ee30493d-1eb3-48cf-8785-31ec016c1689','2026-10-29 10:43:09.606173581','2026-10-29 10:43:09.606444455',NULL,'b98145ab-1497-4ddf-bdcd-9b12c4c141eb',1,'2.oRE7wPjmr7sH690HD1busA==|wCPzZwrrnlWnxPKz40LFGOMgyUi6FeCbhLVMN0Rgego=|afHqAFDKgfIPCSj7dLc8jccuS83BdHV/1l+/Y3Q8+l8=','2.UrV4MX+siTMrIXMNtTBZIw==|dzUongwneyAuPUe9HFwNW9AjnLnfK/7/WBCQjAEXXSI=|8F89xtQxnU9BGD92o3nwS1k+1p6b8Cj5LsuZicMeU+M=','[]','{"password":"2.WzvEgSbuwvYBhm/SxucnBQ==|jb8v0S3/G2NMoIUFsWiPPoVFH/8QBqeoxhJdU9QzTEg=|RCydxikMBiLWpEwEUl6SVbo73agPKxZqYuxMt7ASHlY=","username":"2.OdCKzj8JH0vTHgpmQuR5iw==|bg1DsuevBJsln6E5ZZC5RA==|ekb8ENEYGaE9F/HaBoXqaVE6rA/xsFsFyTORoDgNVE4="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('580f33fc-5ef4-4d32-94a3-a10722db9583','2026-10-29 10:43:09.607313662','2026-10-29 10:43:09.607648370',NULL,'b98145ab-1497-4ddf-bdcd-9b12c4c141eb',1,'2.aSAe6B+Cywy/UxdPDOLnPA==|Sw2u38qB4Ryd6Y7jcNBPjrpOrZHPQ3vbh/T7EhQsd40=|NHXap/zJF/0zmw38fUEEy5qmnWOlBG4Krid8se23ZEk=','2.lIG0qZ8eGKz46m15IKiCZg==|wWS86DfesBCw54N+QRslPKhz95LAeflv+619ScQAnr8=|EhA2m0Q6x8pH6hiMwTtC5lmYY4fGDWw3CzUQQli/ggM=','[]','{"password":"2.nW1ofmPzZDmBzqQNVo8f9Q==|XR2u+UbReZXIf5Shq3TcEsjVP/AQLE53tQBLDmHt6rA=|1kbZcFI3O4tgV3XLL5NZ1DYWioKDX34QGzraMrZkNAA=","username":"2.jDCE4TrlmghWi9s+7LeRaQ==|45N7732gH69PYSYdvaZRMA==|TxOXcvlZg8KTWdliEhkG7Y3Zr5Z9OvF2yknv845WUKI="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('69ba7d8a-2774-4cd0-9469-f5b31928ec77','2026-10-29 10:43:09.615014068','2026-10-29 10:43:09.615433484',NULL,'5bc79d94-9785-4826-a53a-629947978bf2',1,'2.V6yDpXg25/IR5+ThbTJ6uA==|kXZ8eTREXtauAaBRmExpZQtabvcMudGQfYZgprhpkpg=|nryDCdXa4b8moD21Z95kczZbS0oYjJhXlwpSz9Rf8IU=','2.JkI2/Ykx1E5KJkh+ZJrXRQ==|mwK3qofql6JwVd8x0No6bQV4pyThhJrgSXv9so4ASQA=|XeXpp8DXQ2IsAXi/UTsadAEHEAfME+M+1CRzfnq8tQg=','[]','{"password":"2.M9NmMPQfMrc++vCTH2yV2Q==|i05bjXQ6GKjsyFivLj2MJw==|p1Jhxlz/y+acljnmNz5Ug5naKb3z2kN/SxNwxiEQNNo=","username":"2.E+zDYbhh8+cPg0yVmrHyGw==|Q4jLL2EFPTYSE6WHP3O/Ew==|Dv6PrkfWfdu1D5hfSVJ4b00tZYQECjnzR0q+zRdD8CQ="}',NULL,NULL,0,'2.rsf40emxKYWxj8m1l/SGPw==|LldfugKGOuxosAJZgs3T86ZvojV89D1T42mz4GGmnF1gpx5f63Z3+DOEcZ5trs2UT6wupZ8ZTruQAotjWOyCK2i5ITBkiWaNTbWWTq9wfuM=|/qRJp9E3qZbhNkt8Lj4D60f11QmJ8e0zfjx5bW8qcFg=');
INSERT INTO "ciphers" VALUES('167a6544-6e3c-4f7c-92b6-54b8d49a482a','2026-10-29 10:43:09.616348649','2026-10-29 10:43:09.616659149',NULL,'5bc79d94-9785-4826-a53a-629947978bf2',1,'2.4rFw0LC5WV0I9C1zkzoBpw==|Ey76G8rNsPN9tr59hSwJWS1I8PpIfbNQ/cQASe4TZ3E=|+aAKEEH+/e2eWCqI9VosYFGwC0Dp9Y27NOIhqnARzx8=','2.zLvh8Kmae7NqRXIWfPIMJw==|qnJ16s4Y8slAlWhYoXaLgR76/do/vtwW4lQhCowmx7E=|Y8KeV7q+U9fBqEnXLOcp7U/Dcl/kkq2t/vAukWfs/TE=','[]','{"password":"2.8nC35mBLltRQg6Qz7itfLg==|uaNpC3JIYAak1jy0KkbFRyK58wKZl5qwW3bCmQCLgP4=|ieeimPCJXftWjUPF+iyE9oBZZC0bvTOI0UihrWxv2nE=","username":"2.Wl/ouRmJRFI/ztG4qRhKNA==|pXwNo0wDRJwVnmc9CNiLeQ==|oBA5AiVrIMwXBAiveqLTyJZ8TOKsMYe8PnFAmZki214="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('b1e1088e-1859-4c62-b259-ac91a33c927a','2026-10-29 10:43:09.617650189','2026-10-29 10:43:09.617934897',NULL,'5bc79d94-9785-4826-a53a-629947978bf2',1,'2.ho4C+EH2NQOlhZ0Df+oYHg==|uHeHPbm+sf9yjkBX+7kis65ab4Ez71ezuorAKDuElW4=|VqnDveA0OkcrPzU2lSVJ5iEkxP6DbNBGRog77CTpBlw=','2.2wCIzS2d59zTCPI88K6WWQ==|lO9qgRWz9Cgijwr9agUvqlLDGltsn6NirL056h7oLVQ=|hdXCBzBqiZsJau7jIGEBMhMket8yhCBZd5WJ1bEVUMQ=','[]','{"password":"2.q1XJx1rLjyyXVlafBG12Fg==|IMhDLWpTWjP+/WmN56TFDgypADxv3UQyjHWoVudrXzM=|P+If3F6lkP80UaWrmbp9nRsazb0xAItxhDnUDspgE8U=","username":"2.5uNIL6k0dVG/UEzfpkGniA==|mu7R7D93didCNYGUX3tB1g==|hDFQBJNllSFe6ywulxOfnnhHV5ZAulktEcAmsISEyME="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('7dfe5701-4d63-4c3e-aaf8-299f71ef1b99','2026-10-29 10:43:09.883876305','2026-10-29 10:43:09.883935513','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,1,'2.aMmNGu/PMffjLm7W0KbtSg==|niTMshjqQnFXe9/lJyWCdw==|erlXPjhkwBDSJHlBrPKOpkEnhrtpcbDcMV4VUV04Qa0=',NULL,'[]','{"password":"2.FRA8iMKHZsoxx97Mi/UfIQ==|WeAAI0ZlXtLdPVYBL7D3Hef4+u21i8AMObSRbcJoo/8=|YlH0k/kbpEIHWVcGAB0xHdCXJ9w/tdYjQcmvm8SWqn4=","totp":null,"uris":[{"match":null,"uri":"2.QdgHk/UDcWAxR5akfErD3w==|84um9PTxMjVXvSE38Dn0RIHINEIqkVrIdTYz4ZH9dM8=|lJ3M/zyhefOIaT8sD7wPcg0LWYXGDh8zxh/Zq8D8ey0="}],"username":"2.f/P2fbWsIqO497isgxGgPw==|Ckcb0CO+rX2+b4eMPNW/KBCKmG6yuCcn7G9g2dCKcNU=|cClRPqtdiYDe5AM6rZeXvqnCRbrlruYBfAE15Tc66RI="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('d008aeed-04ca-407c-9962-6394bed6292e','2026-10-29 10:43:09.885044803','2026-10-29 10:43:09.885100136','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,1,'2.D+qqQa7X6AMKNa+NukrEkQ==|/STruXMtBYPgtnCJBvaWRw==|Q4awXGovknqZ2fmmwvq2iSVft7+z2o1xq+4Qdvf+Xfk=',NULL,'[{"name":"2.Cjz5Ose1uYxCInHaNkUX+g==|/+qEsR+BAh7NKhDRtxRg4A==|OfRswYOLo3bKfp+xRbeCp6MQ7t7AlJLmq74GXRSe0oc=","type":0,"value":"2.cCofriZQ1StCEdfUCL5Vug==|8zoJXO4bUCZGZlV3eeABNw==|phFd7SjzmyHay/jtYn1KbhDywsUi7Hpg0U4oVLIPgKs="},{"name":"2.kM0eRWw2JMgW6+4/8hhcfg==|YJ+/kaoOMdN3rIWaNYKCRQ==|9yGlf+EPXofStDNigkFElbfpynw5t0+O/eT8Fg2jEII=","type":1,"value":"2.jGX4CJeDwsgvhKtvmDqmkQ==|VS+io4Bp25fla7bOaW2d6w==|lnOgKI4wxdT/HnUeiYJ53m2QikUSsp2nd0SYwUzJIqc="}]','{"password":"2.8MiEj6GcGN/veZ153BEJOw==|Wm48jTbCTibI30MSOjwqxg==|TnnYPyq9e95PxsWBHrd0WCbPXyHcFHFTbZqow22lzqg=","totp":null,"uris":[],"username":"2.PRdu2keQRjoUNJMpbhjXuw==|XAFTC7E8jJINsR2g6kopCQ==|bEhbROl86oMPK9pSErTiwfexMw8782g/SnjyTvhVl8g="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('9601adcf-6a73-48da-bf3a-6e3404dae837','2026-10-29 10:43:09.886126801','2026-10-29 10:43:09.886175843','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,1,'2.n9JZxjTty5pYcpUPTGbUQg==|0xPmOQmSMvg+KAbPbKrPEw==|QZKHPSLmCs+C8iI2KdVx1S7hBaHOPMJ27Cv4GTRBtu4=',NULL,'[{"name":"2.ibKjUDd8pPrYVvqsc0FYCA==|AOFYt9iBZ/T9TvbpsA0O3Q==|rUY5nwtfTOi1/ybxj/8kwsc0Gjdz0FqFZ7p4pEMC890=","type":0,"value":"2.ERWzwir5MqKl5MECiRhClA==|ef9EFTd/PSA97LF0G8NNdw==|yJ5OzkB/9Jd3aGaoDC3hQAqbm/zmxKW+5rlvi1WP+xs="},{"name":"2.P1br2ku4OdLImdhR23Syrg==|dkPsL+b5ZmOzRdb2DQIM1A==|LlWJEnqe1IvCgsZozTguwDlKxtGRssyAwdtxC0kBmtk=","type":1,"value":"2.0OsH+mr0EDmcBGBiry+2ZA==|QrivMgWoRFaqIxkLz30dsw==|RuLD9videUKtEk0g5pk1ePBrdJ7VIpoxZnSphWXoqN4="}]','{"password":"2.zVdK0GUpSijicwx0KdfF+g==|zZ+IVF9NUjZtGUxWim9UEg==|HagczF38d8RimxE9XZ2Ot8U2psERhwLnGquKyenchgA=","totp":null,"uris":[],"username":"2.XVBJsBPbwFSMn0lAx1zpNw==|Ud8Y/3US7FEOxhrllWsVoA==|pSok0iN8NPQ6glMPay/TxtO3BOV3lH5+llZxxuVfkDU="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('44b56808-9344-47f8-9225-e118bb5129c8','2026-10-29 10:43:09.887187633','2026-10-29 10:43:09.887243091','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,1,'2.t3twFFqT6T8CcRGvylpKwg==|Hhnf87vpl42fukZYL0IGNu/pqLP8b9OMWj3ocT/zQUY=|QYmqTO8UH5RgRrucIxoUaU7IDwrG5l/u0aDkWKhCzTA=','2.OFkdWotIcox3fgucA53ByA==|gRJ17nD3KfsuBRvdAGyLZckyqQtyMOjhCkE+gyrFg7s=|gwi8+i0CIGzsnhwBp8OAcgZ5qCmu5opkT9ckHm3o644=','[]','{"password":"2.QYNcu45zd0vaPFCEaKx4mQ==|4LlHXV8zYdqklYQdXz7VqxPGez9kBSufn4M1c+2MJ80=|yUoJeBi88sZRPCes0QRgDK6SeFiELx4pqV6L9rkJntQ=","totp":null,"uris":[{"match":null,"uri":"2.qiXa/db/oP3YQAWx+IQ+qw==|6o1TfHHMO6BSGiZEmc2IFNzvzwRabIXVwZnF7+LbGR0=|nr6FRubh6V6mQt1zRjuLeyGUyEx8v2EOClPHHy789Z4="}],"username":"2.UmAaIEBnAu6H02NG1qDwRg==|ybnP1OpAqwbRN/nF385YagFRv1Bd8m3SSasVk7s5vK0=|LAF48l4Y7gP379Y9O+7i7ZL0qE1AvECoznepdOR/99c="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('f01198c3-e29a-4441-b2c2-ad1324517816','2026-10-29 10:43:09.888197548','2026-10-29 10:43:09.888241548','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,1,'2.WD4/raCRK5SP7mYp1P0gTQ==|aIRK7zMLdCJAiadYkKl4Fw==|ZTga/qyaQ32jNJ+r/HWYm4HNEknLHVwPedCK0GdjYDw=',NULL,'[]','{"password":"2.OyrMRUJah+rX36rOsXD9cA==|BMh7ET8FZjy4u0sURF2GcQ==|K9b/JZoAiUHTaZli253w4/iRsuNvOR+oCbsPcIDcc7g=","totp":null,"uris":[],"username":"2.fiaG/M5+azz3guqRNOJF/w==|ICRhxVppRIXP9vD/kkeECQ==|2CkywXrEoUPL6wGJn+k11pvZpo2YqVca0mquw1J8/sM="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('1b9f0e62-d542-425d-a64e-8b3e87d5704d','2026-10-29 10:43:09.889254839','2026-10-29 10:43:09.889321005','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,2,'2.1CxwYNlAeTRP5EgEnfMTYA==|3AGiAwRrc/3fWcyPTXz/gozzgoSFFxOaIVxdw+3DyHo=|NGHoIlOYdVTvTKrxQ6KCUwkS1QqCQigfX12rgNvU2yA=','2.MaCN7H5PYxnI1wUFMoPAPg==|zKIunUlhyF0/ovbdncHzYo6qBR0Vbn5aqv7mpRI17cQ=|fFWYP6/s6MFRGZM1/Xm+N/+EdabBm0gcActzBzZ777U=','[]','{"type":0}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('e29df966-1f99-4b31-ac78-0ad3f6b65c31','2026-10-29 10:43:09.890333587','2026-10-29 10:43:09.890397962','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,3,'2.BwDktfOXUo7mOr2JLKvouQ==|vyLdu96E98WQv5FOkTU5IlcUVmaAK8HCU4uVhjT6aGc=|+kb38W047BMeUZA8sCnMKgyZLQTtzU+YU8y7bmKz1gM=',NULL,'[]','{"brand":"2.c6YdfU/b+G5vXsKrqfFXEQ==|Pq00MzNspigc3EDYUTCu2A==|TxlpC8RKeTi82QipOt8SJfAxPwpV9JQEThESpc43LYE=","cardholderName":"2.9BerbWYEmz3kV3c78PeXag==|CljD9AYqKDe9E2fX6bv7vQ==|+5nCMh9WzXPf9pbOqr6PxjgRX4hIpq3CxjgVJKTuHw8=","code":"2.0wGoLAf7yi6Y4tj4Kkg+HA==|I3PPmBFA0E3Vywc2x1lTEQ==|xJpLbUf1T/zAbpP97pqcqppoxzvrcmAXp57pOMPU+nc=","expMonth":"2.280Qz3vyHyuOeCXspXDN5A==|lFsAoG8Ry5fNq9JhZH8Nmg==|tFSacavAZOJDh5Ro4/UI8DA3BTveH0JjnMWZ07t1X+k=","expYear":"2.Vju0j+cBDMkJoco8GsJymA==|14ONj4pvcEgRidx6WP6ewQ==|Xp2FiR+A3Y/1W+VDiGqrIniYsomTRjOXfwYyzsZP7xA=","number":"2.00dsi+/pRSVm/yGnHbeO5A==|ymKTitUUSV+ExQRB4pLVg0C/B58rQmAJP80krpq9aJg=|u9VJi+nHMnxgfULYkYy+5F3TdWR9uPlw+sqlVDGWpvk="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('5fabecad-6b3b-47cd-9b70-faf9ec32f41c','2026-10-29 10:43:09.891401960','2026-10-29 10:43:09.891463127','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,4,'2.D51YOFsr1+lfTtFEwaiDrQ==|ORAXsYPriH2M2wRpeUKKKg==|oaGl5lHseP7uK0Ko2wc2tByyjFqyzZKSC9MhU2bBa8o=',NULL,'[]','{"address1":"2.9kpV/in4RYfJ9F9mys3L+w==|ah+PfGsIjnygQiIEoiHEoM7JoR1JG+2kHoRfOIxZvv0=|paK1o3Y6VNfuPRchz4dzRSidZVU84Msqcgy8cVsFbKo=","address2":null,"address3":null,"city":"2.L1+ZV/7BQT7aueXA+g/1gg==|X293bV6sYVk1FDsvsUIm7w==|u4AcTYeo2V1SLV6NMubDLsrAZCJQYLFaQuyIlKk2Yrs=","company":"2.9BP0vrSktLlub4jcX2aw2A==|oNo1BQfVu01goL3tEV/G9Q==|wqSgBRdidf8d0PlOHJ2GX6F5NNy52PRNW/Z3Kmr5PFg=","country":"2.v6yAWuVb742AT/4AezGzUA==|P+WAdHT6xNZLxeBoABvBuQ==|52etVRNzyS6/qO5DeOXyCR95soN0xPA6mn8tkrJzeYQ=","email":"2.LdVoJeSZrhE+iuHjFxtfcQ==|fV8lieXIzR2jw69/7Fs19luYJYpCFdb2vDv3RptlN14=|6Y7ThTYXy0ozIQkHI/G3Rhzkes7p6Ap1BrJBseWAYlE=","firstName":"2.372zRzD8dlH2sqjmB9RZ1A==|mOKzUaMVQ1wzmk1swEz5Iw==|JjZOmcCinDPUPmn8A4f5W6urq/tit8qCoDSo/m6Ro0U=","lastName":"2./z4308l1slcPus+X6bM5nA==|9jfUb2WkBj2JLaSGLOwpyg==|6D9Yrd6PmmuNCBVWXIObYCraVeo9bGOXC8JS+azZ6iM=","licenseNumber":null,"middleName":null,"passportNumber":null,"phone":null,"postalCode":"2.0GEFnGURhwp7vSQg9PIfbA==|OEpZcJUQjj1Ox6LLdRakhQ==|+h3CQOlomXRTvO+M4JTECXQdDsOfqe/M98IX4Wxxf08=","ssn":null,"state":null,"title":null,"username":"2.aW8bfCsemNBAGyeDCl4LzQ==|a2BFp0i8neOYZtNcaTUQQw==|MyrEHxkgB3rmkL41JLW9Hz+7u2ibrA+B+ePmoeMxUFQ="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('5f7950b9-363e-431d-8b37-7cdf2caa6b3f','2026-10-29 10:43:09.892467959','2026-10-29 10:43:09.892530750','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,1,'2.+nVnpzSPxRXL1Bj8LlPmiw==|4Ct0a1W3IMMZ3liACMHq9B/7GSJpHGhwmm5qg/YdGvg=|caPia20Xx3GE3kzUYRnObKzidy5Lx26gCCsHmUwZFaM=',NULL,'[]','{"password":"2.9x3ibFChJ2++2+W5qBOfgA==|IaL2TvDWL2ToZXm5xW0sPcAzVMWSFkaXdp5JjoJF0qs=|A0ieAA3q71VAqfaItKqMgVDD3p10mQKGE8OSX37vkSM=","totp":"2.40VkCOQvHB1jtFpFz8yPuA==|zxAiX58cXuTWEjP+GJuHloV9wzI96MpatJnkepdd6kSsMX6fi9shUQ0VGCpOKtlz|t+40znp8b0h8eUwDnNyFtRukfnHk51MMRq6BpryYSsk=","uris":[],"username":"2.UveAUT5uVgGcdli+IB/99w==|eomw90rZuKAmEAnl4kQI0gOeD/OTQj6l1OUfqy5a2EA=|pMAN21V1F/EGCuMHbYQcNS7eckRhgUG0eoXM8K901g4="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('f7bfe7ac-0cfe-44bf-a7ec-906eda971c68','2026-10-29 10:43:09.893449374','2026-10-29 10:43:09.893501416','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,5,'2.uog0K0ouZMyJAYwjMeR7mA==|vjRkX81ANnzPvpl6mLRuwT35uX9P85BSke1rzAMKCEw=|Ep25Xl8uRdHOu/eF6nhlK1TwcKk+WmzjNJp0QKJPTvY=',NULL,'[]','{"keyFingerprint":"2.uQRFHouBNZo1RnxG7ZCFNA==|YQx/rwfAU9qgSyz2krzJSUv7Xq1P3g16uI2p5TdfgTLMur8vodyBQn6GY9/8YhquW1FbyitcQxw8zZOIzv349Q==|EYlXFuwfzYwePWH5soVXgyKrMAyOTFCBd3WA1ONP7rA=","privateKey":"2.a56kyjKTNcdCSUl3CBMdlg==|gqqU5f+35lPWz6Md45S4B0bnrXMJx8VFGqjUxxxk+LxRjjxL23rXJVzvx57/fQwrDxSNFAa/308DluAMqTPYvja14yddhhT2o8OZxfHcl52Tkb1vf4zerfj5fzpqI3YXCxyjS08c2ifdEPn9lg3bRF8CqHmvAIc4wpRW2HgCxULX8pD8dk9YmnMBxhYaF/nsJtUy3Lw906+N9Z8GwD3PolovSrH0ysTFUeY0apQWsQ16lR9ccs9Z8GhIpO7V5MKEIezmol5P9jBHCVrt9+YjtNHOUvB7jewFYp9Jm7C5Fiy8CBDtEU+K3XdKrXM6gtWuNstMmGx+2OxPnecre7ocSAYc/xvChIzaZurwwn7Liu0l4ZUfjNJhNTx4bSVMSsF+c7hY1hJ2+pfVRzOzJjXzwfvM02nI2XsxcTKeBjYbr3VzSt+wcWHs9lDg3jxJDuwLMTiei3QaJh+1/Wd/zJA83rSaBTgnrKdHIPeaWqBlJgIhhgfm6ByCgG/8Z9F9P3UPOqcVkSxzT9GlboUGQFAOvw==|sFuh5OyYB+OuOxgCzTtr9XhaIM3ZIhExUhYnYDsIeao=","publicKey":"2.t3ES37Sc3WRnbzKvlINFhA==|Wn4Bv7/UF74Fbi84CqIzaL1KTC4t4Ki5vdwBVXmLelXWcjHNYcdNiBxy58MLBtJdAz9EMb8fItjJQl+GunoQhbZqgUabPVcVnPLGcdnjDTj+CeIUO3ARqtMSFf2DVTE4|S+qOEn5hdM+WthKGtIDPd5vL0lcHbiny75Lge8dDNN4="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('bd056fa0-67f7-4017-baf4-46a1048f5a40','2026-10-29 10:43:09.894466123','2026-10-29 10:43:09.894510164','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,6,'2.YWY88wiuGzTSj6JVhirP+w==|JU6aLlScxVuO3TYAotDvxRKNhky/OLkEWF1hkGt0v6c=|iXtA9/oWf3sNiiHiwVxoUBv6GMAJgp3m68TYF696vu0=',NULL,'[]','{"accountNumber":"2.u8AUg2hgdh2i+3xOI5dLqg==|QGlHZkjoULn/ivTajxDLmw==|mfTQ+J/+o57DoszXEhj5gh0ooK7XZ79+HaxdCP99F78=","accountType":"2.oq/PTvAF4Zen+AF1r5oBkA==|d7YrizVcz42H65QGbXq6xw==|uFHsUj82PBqcqJxpyCzkKUae4sVPbFwRLdN/AR+HlAw=","bankContactPhone":"2.uUtcnuJ3GaqV6B/MvpyTZg==|SudNg2fbJ47bJ0Cm759lNw==|UAxcw6qiu0K0QOSqjIKNIq0N+HsLU77XPvq5hMUCunc=","bankName":"2.VuKcXYTwZhURN8+rkklSIg==|zxR6TeR067mSngu6ELAzDA==|+lGCspTrCs+x5ArjLGabgjzt8crL9f52+9iN+Tcycoc=","branchNumber":"2.tG0S4RkQsToWXEqdE4bNTg==|a89UF46DC3luz/9J2XIvyg==|blsKWC8ejimRHBj0exTrUJW2Rq4ygRBPt+Lnx4ZEA04=","iban":"2.yWq8cQc/Lcf8DPgHynLzBA==|MDl4RHZ3MFrHOroJ8dsH0Q==|IH7d1q+WPpAq08gOKJnPxWTVVJVB7OYHm6tF++lQRLU=","nameOnAccount":"2.6WF6RRbogFBwgcG0JaRP2A==|HBEujPEhyEhdZ2AFZIqHFQ==|fWFtyiDthT+0r9YfvsDznVo66FfT5oAG4Bt9DrL51vo=","pin":"2.3VDHN2SQZtYt4+NiO3hIYw==|f6+BwgzppZFVXclaNPrzkA==|3UhJ272TfkYLiz1OdUYyQD9lGSHJuiO4W+kejd5/vFA=","routingNumber":"2.Hbsl7NMSiN+RpJelQVK+/A==|T0RJFsJ16N4l+utVoseHmg==|etVVKbnMbo7iRni1f8RZ7e4pSvgPTGlXnvh5a5yNsBQ=","swiftCode":"2.gmfXa3uwGMFjSaQ8DmlGJQ==|KY8C8alXdmi2ofw0Idww9w==|YBuVUv+0HtLiAo829QyaaC85+rnEm4AyJj+dy1vCpmo="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('d61763e0-139a-4bb2-bbfa-a7a92c1b8cdc','2026-10-29 10:43:09.895431996','2026-10-29 10:43:09.895468538','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,7,'2.abZRi8ZgU7dWtaZEJQMqsQ==|L1SG9eLi00rb5ADQtIjoUSSA0xFY+68iR0VqiByklsk=|I/HIFJg78bzKYfX0tq9K5/SsK1aYp3Im4VpuL4DVBwM=',NULL,'[]','{"dateOfBirth":"2.37qoIPty18J+RHho//HRJQ==|flBCbBZR0+W0FLskyRiBRQ==|WP09hr4dUw/cotKYohLxaIhPWk6a+o/2jt/syUGzgOk=","expirationDate":"2.pBGFi5A0PqxnK9mQMz8sTg==|BDu7MlBCffJzNjr8F/9kMw==|/nFQY8revCI1Je19w2L9kk30kbzGZuXV5SygUllPcCM=","firstName":"2.GlBtBKzNsns8fwKBnqrP3A==|Ln61m2mPd7wmCmRYJdCXrw==|6pG/tY9ImMjqKoyM/Lhkw24vu+TsZ8c7JoQy5eleQQg=","issueDate":"2.vt+v4uZtenakXgShWz1Kmg==|cxhC1xA5ZByIVF3Yxb4b9A==|fSlj1/GZvp44u3w23B0eyYP4hN/Je72hWwCqKzMhFVc=","issuingAuthority":"2.ljRszHis851n+UnO9IexCg==|PFIClofVvB6hz0y/zN0yBA==|r1tgbmlgfsyIsGNlWOKiVIuoWmnKyJHDOexe02bezuE=","issuingCountry":"2.LgJd8pdbxUOa5S4PnZ1tyg==|ltbMz3Z4uVF7QsVAkW7t/A==|K3FJnRFiSJrnUzhPTCrWp7PGVLj+mFR1UbbO2G7jRyk=","issuingState":"2.TVkIVxZYG5TIx23iM+qIOw==|e4/3e8Cv8ZTW3lta8SKytw==|GzDMu1wZyf8zE6wz9nipDojC4jkfiyxcUM9fUHHMIsY=","lastName":"2.zQXeMo1gg0eVoOSRzeMlBA==|/r9+ldY4oAOaKbJyGk/vcw==|sAUPu5APRhWAff3dLafWbwGkgrOY2LBw0SDtsWL9RO0=","licenseClass":"2.ElnRwAvh7bLMOhDF8kCW7Q==|xluoRW7MgyhJ3y/E6HrDkA==|1QR6gmzRN57KzE5ZjAIF/ZcLxjingHIAKd779YqEHRU=","licenseNumber":"2.wzjgEvE8n+rWnq+GG8zMAw==|tTyGQUPyKuED2V94woBBIg==|dJqC7YS+AjHqyG5lABW7SL5bi1kF80LtGK0N57kLKNM=","middleName":"2.gIE2ApiyemT5+hKi4wFxqA==|wyjCPZA1+xuGJhz97f0pCA==|q7F7Kq/ZUXBKtXt/fj3dFfUlXeVj3NiDI4vDVYDMD9M="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('8253988c-f426-4250-881d-c45ca46eca9a','2026-10-29 10:43:09.896394078','2026-10-29 10:43:09.896435370','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,8,'2.geA5FfYqVWyZ/gkRKFDOKg==|EzwWky9zQJAFBaZWcom2ww==|mMFX08gSt5ZmJXeQLRamB1ugti4AnWeH5pvOEIBRMuM=',NULL,'[]','{"birthPlace":"2.UcJpV11eEciCBLFp4c5fFQ==|Ex9gGYMajqQZ9jAJzHjmaw==|3eruKp6+9ZMBl+OAnQkCoKoYR4e8N30tW8zh5uEOZBw=","dateOfBirth":"2.xbuyqO+KJ6hSKmzTcw+r2g==|DLREV/UPgbNFYnyTf4yyMg==|5+ZgjiPh56kTlGS0i2oQ1InWrC8TJr2X7pL7ERwFrEI=","expirationDate":"2.y+tGEG5OlkfuB9oy8qM4bA==|6VDK4NWYx7oW6KFRGHtS6g==|ebi/lj8j2EAV+dacpL/lu51VR16sEE1Ohi/97SfHCxY=","givenName":"2.te3pw7O4b0Ha215HYctJdQ==|ECq1bFkBRs0c3n1AuPR8Eg==|K4LudVnlX5oDGtjhlW3LMdJljqR8cExi2rdFdx5Xkcw=","issueDate":"2.dYl9XhaD+p07FnaewE1b3A==|d38lJSc0UQS1cpVm/2EoVA==|k6AtlpZwmQ5DnRnYP1WoRFL7HhemOzKiqrwtzA4IN1I=","issuingAuthority":"2.9ELcMYE2Z9a1ESlHjZaHuw==|ZjWNzb1n9evwPS5yM6Ur2A==|0RxV33nW3MIl7uE5+F8nadT0nX7v8jbP3hkztq84nC0=","issuingCountry":"2.AmrWntqtcNZe7Z81QswxgQ==|b6Mw6yxvnXxKO6cxCy7G8w==|C7sfY2RXkOKkirBHZE99XC/DiXpnJLzjzAPqAKEkSG8=","nationalIdentificationNumber":"2.Ay+FlQ5My4zZvzezqcKkUA==|khEYsXr0E/XdUomC0IQo8g==|sdKjP0jIEbzwbBIgdMAMTVdZsDKfTzjRtPauM4ycUL4=","nationality":"2.NEimdNHMdEnzqJgNoDA1yA==|CRcsEYrKGI9QAsaP2Piz+A==|aMpvDs7gmrYRj0kZX/veBLYajOQHfnh7m5bisPTCsfM=","passportNumber":"2.jDw7QutmxBRrdi95VjRcSw==|3XcGwai4joi6/W1KYsxioQ==|Lq0kJ7ueyztDfcMPX2GE66R/EJWnleSeQp0ie6lXhOc=","passportType":"2.u+OASvP1Q4vFM2LqHOxlIA==|n6jP5/9XE85KhSndHz0u9A==|6OshxcL81v2iCC1xAz8fKQ1dnXhFbCTlO4oUqi3D6Gs=","sex":"2.vvrc4aIcDqfGTIxIgv5fuw==|KeG2QDbWlFacFA6yQn0xeA==|B7n+//yl/RMc0fr6pY1kYtv7aHJKRFoQ8IP9j6X2maA=","surname":"2.lPswNbB1p801wzgk2PquiQ==|htAQhsYn/60iYg33wqJU6Q==|lD+kbJdJc9Dyw/qyoNneOps7CbZ4fJpmkcgKHgudgxg="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('770581e2-483e-4f09-909c-24b9dea8ec48','2026-10-29 10:43:09.897288619','2026-10-29 10:43:09.897342160','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,1,'2.Zog7WqDPBvfRSCbGPT+PyA==|fNiv062FX9Xds2jGPHSN/TDSm89ge41Q8spgADQi2qI=|5bQ3vicEJhe3TqUqhPl6CRi3Whg7qWgmFpOG2d+Ziqs=','2.NIOO/CK1cHYoV07XbxwgqA==|IpZ7CX/AATx3W9up3VCV2aEyDvWB47aEG22LAOdttUw=|aOplWtdgkE3Z8TZ2cSYcfddiVgjwmK1B4Xd5z+wwOuQ=','[]','{"password":"2.BnuAJNUutx7WYUBSuG3jcw==|0eiJF4F8CJwKLXTiuXvBw8ofL4pbfeWMPatS63YQIdE=|sVG5OYh4ClG3mME1ZqRxJgBosqwnEvvE5ZhyZuhF9vA=","totp":"2.1WYNInbiLywOYuewd0iBHw==|ZSUmHwZLVcxBlEqX/Fnv3gfeWYU3qOc7sWfyz/DZ0pg=|5MNXobZzzhr30ywYFfI7bQ7SH3j4/vzkselJsH9uBH8=","uris":[{"match":null,"uri":"2.RMkgYHAtd9L2Dd8uABOTIw==|ttShqTmRSqfdLxuvUlgSQUbcEhI9Fj3hVDJXeWtm00Q=|WEI+jremb3dmVdUH5/upxDcgd0mNi1HtC2N81fk9KFQ="}],"username":"2.rInUNDLY0XCufOvgHeZtPw==|ndsgpGASa0p679sPmoxkApYx40mkNxfLHvPfWiur/Cc=|io8gLAcgNMlIFXdnptSsXUjbdaED28/SUbWsGX0AqCM="}',NULL,NULL,1,NULL);
INSERT INTO "ciphers" VALUES('c526a36f-1943-446e-a66c-c3ad4bc09b0d','2026-10-29 10:43:09.898262534','2026-10-29 10:43:09.898314992','de1dddaa-3e40-409d-9eae-cdc73ceba1c3',NULL,1,'2.YX8sXNlzDjCsx3LLwtR0oQ==|+r8IduWX7rXZFr5LiazHV9VNA4iQ7feGnPOP5XQ3i7FkHEAY+oTQMPVv/8LAcXrL|YOHGCPc7ho0DanK+C3x4kzscrib46HpZkYxNoB9B2tM=','2.0X0VidCFRl9UaJpoTF+24w==|1OXTDbSSR14MOO7sj+/8Qb5Y2cN8JNe1I1S37kEQekY=|GGMoop0tmwO+lT1sAdaAWbBQjubxga5zH1pmm1D+uMk=','[]','{"password":"2.CAqnvnmbETmYFxsPA9NsNQ==|gkrXa2ThL+t4xtrwFJsSXpzk0wyIndmKjj4YxgeC3Vo=|tfvTP9e+YD1LAijfPMKBTDBdtyXwG7MfiwYIGAQAeZY=","totp":"2.CWVT5M9f5aeJhPbY4SY+fA==|t5UMnlNHex5/rHgPGgwGT00T1WHqn1B3obAfwTgz25s=|XWwC10LgpTK20PJUbP8qVHWavytoGX3NjpayrjVz8Ts=","uris":[{"match":null,"uri":"2.6Kvgv+HCOsuQX1c906+dXA==|rk9gzaBqAKZQV8A2z0CZC4FLzFcKoxzZCEMlo5OFEQs=|9DjhBNgahsCP/jSKi9shSR1YL2h4CUCIK4cYlEHKUXc="}],"username":"2.+xXlAVJb6JE1rq6DVsGtSQ==|t97qugxBw4VfTIlbBGDa++Qbz5pUkJilwcBiZU+PYnc=|GYE3RGF0+mIK1I6TfkQZPosDUnEn87D3BTEwUkhZDaE="}',NULL,NULL,1,NULL);
CREATE TABLE ciphers_collections (
  cipher_uuid       TEXT NOT NULL REFERENCES ciphers (uuid),
  collection_uuid TEXT NOT NULL REFERENCES collections (uuid),
  PRIMARY KEY (cipher_uuid, collection_uuid)
);
INSERT INTO "ciphers_collections" VALUES('040a4360-e1e7-4160-9924-f0b1af25021e','204b6bb8-29e2-4ed8-a66a-1cb2948cfff4');
INSERT INTO "ciphers_collections" VALUES('040a4360-e1e7-4160-9924-f0b1af25021e','b0c2e10d-1267-448c-98b4-5401e95c9a54');
INSERT INTO "ciphers_collections" VALUES('ee30493d-1eb3-48cf-8785-31ec016c1689','204b6bb8-29e2-4ed8-a66a-1cb2948cfff4');
INSERT INTO "ciphers_collections" VALUES('580f33fc-5ef4-4d32-94a3-a10722db9583','b0c2e10d-1267-448c-98b4-5401e95c9a54');
INSERT INTO "ciphers_collections" VALUES('69ba7d8a-2774-4cd0-9469-f5b31928ec77','42048c39-f285-4713-a5db-752bbf98ab3e');
INSERT INTO "ciphers_collections" VALUES('69ba7d8a-2774-4cd0-9469-f5b31928ec77','ee6bb895-16f3-4d53-901d-75fd4ac90eed');
INSERT INTO "ciphers_collections" VALUES('167a6544-6e3c-4f7c-92b6-54b8d49a482a','42048c39-f285-4713-a5db-752bbf98ab3e');
INSERT INTO "ciphers_collections" VALUES('b1e1088e-1859-4c62-b259-ac91a33c927a','ee6bb895-16f3-4d53-901d-75fd4ac90eed');
CREATE TABLE collections (
  uuid     TEXT NOT NULL PRIMARY KEY,
  org_uuid TEXT NOT NULL REFERENCES organizations (uuid),
  name     TEXT NOT NULL
, external_id TEXT);
INSERT INTO "collections" VALUES('204b6bb8-29e2-4ed8-a66a-1cb2948cfff4','b98145ab-1497-4ddf-bdcd-9b12c4c141eb','2./ZrSqXoPRF30HHni7ElmZg==|CtG6fAZ+OXCbj1MBYceKDw==|ThTQx5bZVIv1vg9L5dk74RRudM6Rq4ijaCkiXj4ebDY=',NULL);
INSERT INTO "collections" VALUES('b0c2e10d-1267-448c-98b4-5401e95c9a54','b98145ab-1497-4ddf-bdcd-9b12c4c141eb','2.4JLrlfIsxHHhdf2JoKAdrw==|WBzjBMs5V5+NGi/3KnVRdQ==|tLq6ssa07ilp8CUj8YKdS9L3Tw/SI97XF2IjqYIt31k=',NULL);
INSERT INTO "collections" VALUES('42048c39-f285-4713-a5db-752bbf98ab3e','5bc79d94-9785-4826-a53a-629947978bf2','2.0fQN+/viwbjxCC0h9sAntA==|egKOJiN+5g0amM5PLsMROg==|W8XgHlKBdjckAGIHvnL/3D04UDKey98IJpZ5gf6vsSE=',NULL);
INSERT INTO "collections" VALUES('ee6bb895-16f3-4d53-901d-75fd4ac90eed','5bc79d94-9785-4826-a53a-629947978bf2','2.Dry26SPkCCKkDEG5Na5uyA==|Bu4g4wmdJA9TOexelIJzMA==|aey2Ajd9+L2wOJYtYyirqNl2XPGpcBwhdBs+DNIM3CI=',NULL);
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
INSERT INTO "devices" VALUES('30000000-0000-4000-8000-000000000001','2026-10-29 10:43:09.340310673','2026-10-29 10:43:10.105183318','e08aec66-84b3-4d66-96c0-794442636726','bwb-integration',8,NULL,'uUDqrL-xtwDZu0YMV9PiFFCFm4BAjSsRwDDN24o9yQb441Rsrz-N_tbcYCQ-0QwmiXJ7I7LeYMyAUHnKfMpMvw==',NULL,'e2ffeaff-513b-4e10-8f50-76e1d0ca4001');
INSERT INTO "devices" VALUES('30000000-0000-4000-8000-000000000001','2026-10-29 10:43:09.878169230','2026-10-29 10:43:10.726196713','de1dddaa-3e40-409d-9eae-cdc73ceba1c3','bwb-integration',8,NULL,'v8jY5H5TkdwDOEABfQZp79FWqmm2cCh2XYSjE0Wg-WG6yHvUltOaxZ0H_2JqgrBjaWzxorAi8e0WrAMa5B_kGg==',NULL,'6c680624-a9d3-417f-9d51-78dd0868ed42');
INSERT INTO "devices" VALUES('30000000-0000-4000-8000-000000000001','2026-10-29 10:43:10.260859593','2026-10-29 10:43:10.260952343','7d8c9c1d-5539-4443-8306-f49895c1ff81','bwb-integration',8,NULL,'eBHuy4AR7XRkZQZzA2mBaTFZytZwKm3MSaiYfMHBQiw_ZPJtd0Rf-Ldca1dNxuKOlYD38XOpY1qqvHqJ42cA-w==',NULL,'9d2b30af-4d0e-40de-aef8-58478973f7ca');
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
INSERT INTO "favorites" VALUES('e08aec66-84b3-4d66-96c0-794442636726','da2d518e-46eb-4391-a4ff-97c965c3c7b3');
INSERT INTO "favorites" VALUES('e08aec66-84b3-4d66-96c0-794442636726','32a99461-4243-4fe4-afce-ee73c2bec436');
INSERT INTO "favorites" VALUES('e08aec66-84b3-4d66-96c0-794442636726','5276181e-1607-40f1-be3a-40836619ae22');
INSERT INTO "favorites" VALUES('de1dddaa-3e40-409d-9eae-cdc73ceba1c3','7dfe5701-4d63-4c3e-aaf8-299f71ef1b99');
INSERT INTO "favorites" VALUES('de1dddaa-3e40-409d-9eae-cdc73ceba1c3','770581e2-483e-4f09-909c-24b9dea8ec48');
INSERT INTO "favorites" VALUES('de1dddaa-3e40-409d-9eae-cdc73ceba1c3','c526a36f-1943-446e-a66c-c3ad4bc09b0d');
CREATE TABLE folders (
  uuid       TEXT     NOT NULL PRIMARY KEY,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  user_uuid  TEXT     NOT NULL REFERENCES users (uuid),
  name       TEXT     NOT NULL
);
INSERT INTO "folders" VALUES('cec655ed-f5c5-4256-a39c-8bd90d5e42d3','2026-10-29 10:43:09.345420999','2026-10-29 10:43:09.345481916','e08aec66-84b3-4d66-96c0-794442636726','2.+36hvbR/2B3oZ3m/3Jj+fg==|/74CL/MEoL2GxtsgDT42aQ==|1k7ABzcJpUDR3aTO5yW/zqMtr4fS5C/q1tGPFMEHmxY=');
INSERT INTO "folders" VALUES('42aeda2e-2efe-499d-995a-fe4cf230b829','2026-10-29 10:43:09.346583247','2026-10-29 10:43:09.346632664','e08aec66-84b3-4d66-96c0-794442636726','2.kk0RG8u/9QS7OGkSMIJyZw==|UAgcH35HvfB34gs6x51CCg==|OSWAvDciuf3Z3mY1yrpLfcJgIRi1TgWFvIgvVOoUpFk=');
INSERT INTO "folders" VALUES('33adabf3-6c6d-41a8-ab40-7ea1156e4d6f','2026-10-29 10:43:09.882096766','2026-10-29 10:43:09.882154265','de1dddaa-3e40-409d-9eae-cdc73ceba1c3','2.3leJ5quO3CdAOAETWUu8fA==|A3uPaUAiGCScG7uEmTxKCA==|KjKiNOaBOQzH4v9QIe032YPAH3W2jMj/OyDxfNXaIZY=');
INSERT INTO "folders" VALUES('5c37fa65-c154-4d43-8c1d-a393ec43dd49','2026-10-29 10:43:09.882931806','2026-10-29 10:43:09.882956723','de1dddaa-3e40-409d-9eae-cdc73ceba1c3','2.9hjyDsCQaEcgj20EvMh50w==|OP+plEmKTaAg76j9lUUC6A==|3ujjaDwpjud37AKZwT5BRWlba8Rq5DqXW3ugfrxBcBQ=');
CREATE TABLE folders_ciphers (
  cipher_uuid TEXT NOT NULL REFERENCES ciphers (uuid),
  folder_uuid TEXT NOT NULL REFERENCES folders (uuid),

  PRIMARY KEY (cipher_uuid, folder_uuid)
);
INSERT INTO "folders_ciphers" VALUES('da2d518e-46eb-4391-a4ff-97c965c3c7b3','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('56e875a2-5a31-4a2b-b4c4-6dfa656a8cd3','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('5acaadb0-a930-4442-bb3d-f07f5cd5ddfa','42aeda2e-2efe-499d-995a-fe4cf230b829');
INSERT INTO "folders_ciphers" VALUES('68d7d179-685c-4912-98f5-1e90a98ea70f','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('8002442c-cba6-4363-a28e-6dc40902e0fe','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('9b26b892-31d2-4604-9ace-05193dad3255','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('e993b47e-3ed2-49c9-9f45-cde5a757e39d','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('40016660-713f-470e-b21d-ee9b6a032129','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('32a99461-4243-4fe4-afce-ee73c2bec436','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('5276181e-1607-40f1-be3a-40836619ae22','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('040a4360-e1e7-4160-9924-f0b1af25021e','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('69ba7d8a-2774-4cd0-9469-f5b31928ec77','cec655ed-f5c5-4256-a39c-8bd90d5e42d3');
INSERT INTO "folders_ciphers" VALUES('7dfe5701-4d63-4c3e-aaf8-299f71ef1b99','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
INSERT INTO "folders_ciphers" VALUES('d008aeed-04ca-407c-9962-6394bed6292e','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
INSERT INTO "folders_ciphers" VALUES('9601adcf-6a73-48da-bf3a-6e3404dae837','5c37fa65-c154-4d43-8c1d-a393ec43dd49');
INSERT INTO "folders_ciphers" VALUES('44b56808-9344-47f8-9225-e118bb5129c8','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
INSERT INTO "folders_ciphers" VALUES('1b9f0e62-d542-425d-a64e-8b3e87d5704d','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
INSERT INTO "folders_ciphers" VALUES('e29df966-1f99-4b31-ac78-0ad3f6b65c31','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
INSERT INTO "folders_ciphers" VALUES('5fabecad-6b3b-47cd-9b70-faf9ec32f41c','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
INSERT INTO "folders_ciphers" VALUES('5f7950b9-363e-431d-8b37-7cdf2caa6b3f','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
INSERT INTO "folders_ciphers" VALUES('770581e2-483e-4f09-909c-24b9dea8ec48','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
INSERT INTO "folders_ciphers" VALUES('c526a36f-1943-446e-a66c-c3ad4bc09b0d','33adabf3-6c6d-41a8-ab40-7ea1156e4d6f');
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
INSERT INTO "organizations" VALUES('b98145ab-1497-4ddf-bdcd-9b12c4c141eb','DEMO Engineering','bwb-integration@example.invalid',NULL,NULL);
INSERT INTO "organizations" VALUES('5bc79d94-9785-4826-a53a-629947978bf2','DEMO Research','bwb-integration@example.invalid',NULL,NULL);
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
INSERT INTO "users" VALUES('e08aec66-84b3-4d66-96c0-794442636726','2026-10-29 10:43:09.220531221','2026-10-29 10:43:09.618017314','bwb-integration@example.invalid','BWB integration',X'3E5CAECA6F9FD14F9D8F2C5152BB7F286C52E7B30B7627A02CCA82F1C3C5F397',X'6C3C6038581EB4FF7F2380DC1463E570223B445259BE6182E06320F8444B11ED3C5F40C9AFD99BF0309952DF439167C90F1B8942606CA9CD2E7729BC0F39D157',600000,NULL,'2.zNCNZKwHcgSU3n9awejSeA==|+LNCWqqc0+LC8YcsgfAD+QYStRAwrkjeESrPEfNyJvo1kVdmD0mOKppGZh4vfaxY078TiaIA46Dv7QSG/W3vs9d/4X+BRPLuMM4PWRYhC6E=|woH3KbBNRrGL851ocN8tBFxW3ydE9QCfMRgmMfaIxTs=','2.zpJUNzp6AMFtgVqvXkTPsg==|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',NULL,NULL,'9c07e496-9426-4462-8ce0-88781e8e3132','[]','[]',0,600000,NULL,NULL,0,NULL,NULL,1,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO "users" VALUES('7d8c9c1d-5539-4443-8306-f49895c1ff81','2026-10-29 10:43:09.522399785','2026-10-29 10:43:09.618034480','bwb-restricted@example.invalid','bwb-restricted@example.invalid',X'AAE18DD00E6C8226142A7E38EC4FE89082F7268DDAE4F3522DBD3C671DA948AC',X'F6C9579C155DC6D96255774D05E117F94A08DF5E868EBED8B5D3E3561DA323517E065E2394A0805DA144F27318CE941CA879FB8A1F18B5AC5C8DAA47AAEA12C8',600000,NULL,'2.aTmEtMvKRoo9+IYWQvbDoA==|iV6a2be24P8gTM+lF/jSHyaIVkG6ZtuyyyS9M+seB47F5ZE/LBMx94w04/uxGZna7QCxAWCG20OjMEmHlwwaSwtScIOa3RCDUR+errzcjZU=|Cz3O7D7DN1hSXrdSc9mho1smkADfiStzhwwEB5RuMUg=','2.iAJMBBRKC7SsL8WF73gl1Q==|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',NULL,NULL,'42382ac6-0abd-4a97-ba2b-085327df5cfd','[]','[]',0,600000,NULL,NULL,0,NULL,NULL,1,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO "users" VALUES('de1dddaa-3e40-409d-9eae-cdc73ceba1c3','2026-10-29 10:43:09.761992856','2026-10-29 10:43:09.898467200','bwb-argon2id@example.invalid','BWB integration',X'6F7883D4A9E7C31AFC159D16F1B652EDDB7D96523422FB11BC4007411FB095B3',X'BE5DB9A6B622FA83BBB4DB48CE78632B0E36CE2513CDB18A62658ECE6B75713FF25E71891D0FC26B80D0D4C2E531C62C5D1379D81447580857ADB026CFDF574B',600000,NULL,'2.K7EyChWta0r3IecmWjfvTA==|GMPlKw50rgkOqjmvt3A0lQLbjhB8vZgmiWLeR85WsfG13Ld6em3YwyhtBdq1dd9RxNX6OR4wEYDkyS40jf0+JE/Xa6KGSx7/C8jVfvPLjuY=|H3HLBTz5BYeHRJn+oC4PMoMtZEVPOVMO3Z82KrDwMuY=',NULL,NULL,NULL,NULL,'144e2b04-a361-47e2-99c3-e12afb1c4d36','[]','[]',1,3,NULL,NULL,0,NULL,NULL,1,NULL,NULL,NULL,64,4,NULL,NULL);
CREATE TABLE users_collections (
  user_uuid       TEXT NOT NULL REFERENCES users (uuid),
  collection_uuid TEXT NOT NULL REFERENCES collections (uuid), read_only BOOLEAN NOT NULL DEFAULT 0, hide_passwords BOOLEAN NOT NULL DEFAULT 0, manage BOOLEAN NOT NULL DEFAULT 0,
  PRIMARY KEY (user_uuid, collection_uuid)
);
INSERT INTO "users_collections" VALUES('7d8c9c1d-5539-4443-8306-f49895c1ff81','204b6bb8-29e2-4ed8-a66a-1cb2948cfff4',1,1,0);
INSERT INTO "users_collections" VALUES('7d8c9c1d-5539-4443-8306-f49895c1ff81','b0c2e10d-1267-448c-98b4-5401e95c9a54',1,0,0);
INSERT INTO "users_collections" VALUES('7d8c9c1d-5539-4443-8306-f49895c1ff81','42048c39-f285-4713-a5db-752bbf98ab3e',1,1,0);
INSERT INTO "users_collections" VALUES('7d8c9c1d-5539-4443-8306-f49895c1ff81','ee6bb895-16f3-4d53-901d-75fd4ac90eed',1,0,0);
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
INSERT INTO "users_organizations" VALUES('6f1fef70-61dc-49ad-bb05-2d965545a1ad','e08aec66-84b3-4d66-96c0-794442636726','b98145ab-1497-4ddf-bdcd-9b12c4c141eb',1,'4.T/kC5XvcUO+VR/MAdbpQUOPkgg1lkaLIN3lsZMYpSLONMDyiiRXD8Fj4FAak+DdxNpPTMfUOJXzGfsnRDl8UMKoC2w8ZtC2jpbcHio3hIclEPCTOTkllWBLXyvDLg1ImWWeCCJ4OdHZogYvZBDaPBJJey7kUsr1fDLggAHxU55rgEm56ddH01GUiTqqSSS74Z5BATkcaGL0VmEw1i/EDQF9RbMMRGpnUyeXZqjqIpGOrnsGir7ikzRPZ55+fvNu8JPTMi7D6rqmfgfXOJX2/Y1q2sqjjRVXmpbZLHSnY4upgTVs3DQAstmsJ3n7rbsDAy2UKRlAgQGJEGqf+d9J8uA==',2,0,NULL,NULL,NULL);
INSERT INTO "users_organizations" VALUES('a32869eb-60d4-4168-8acf-94f4576d27a3','7d8c9c1d-5539-4443-8306-f49895c1ff81','b98145ab-1497-4ddf-bdcd-9b12c4c141eb',0,'4.TgOfLz3xwzjPNlyv1R/kGwW3nFLzL4zr9yOcDYkXne51nI+d/RFjeyjpvJsCWkVbgTV6yYoxEd4Yq/fZ7eiuUALGYwRe2QRBMg1Dg31muRpQu+ceaTJuy55NhcSSDzJKoT0+PRx45BSUruiTimk0o13w+DBZCi2SfNWYx4GSLjMMsuLJZxw4W1dZMipTgu5hAVon0CdVaQBMtSLFq3p7+SQBsoOWPaDzIVe+mxjwt2+mv4ykRjoKdVD3LEw+epoGiXYdisHm2iAqi9HOgJsLnAI0S7O2rCEzwK0VaxK9Onz//8aU7wNqsf9G61lypH4Mj/WxjPCJL2oVyVwgc7XSKQ==',2,2,NULL,NULL,'bwb-integration@example.invalid');
INSERT INTO "users_organizations" VALUES('406ef151-63e4-4f08-beb2-9de36798a595','e08aec66-84b3-4d66-96c0-794442636726','5bc79d94-9785-4826-a53a-629947978bf2',1,'4.D4kjQnxXaeI9uC9VCA/XuobNqs1HPtMUwxDXwd/kNLEoLvIf0y+BC1W/Z5Vb96zgubfKFkb/rP27pZ0/HsYakEs2qx//ORoN1n/7tfVM9T6sp1nOv0inQvSDF8tBCBlkx415cglXdRYErUas4v4Kk/LWc5BD40LHfVSGfUHFzEoZoDUSRIvCZxn2HlJVyhykoodwXlrm7WJbV9OZ+wjI4Zfde5oTD7dkv/kyCVBJfGNwx6kDV4E9FYxrOimlNUGxe/jivD/Ax34B4eDMiIDENDVerTuR96bqES5yPO7AjY38kLIzUIHHpr0QYj0MFRra5mKzQmpqUBa42R67bGSGwg==',2,0,NULL,NULL,NULL);
INSERT INTO "users_organizations" VALUES('baa45531-0d2d-4c58-8f07-1960cae697ed','7d8c9c1d-5539-4443-8306-f49895c1ff81','5bc79d94-9785-4826-a53a-629947978bf2',0,'4.F3r4+rPp3VXCuz86o3eP5a+NnZDA95zBrrKf8yo1NpQ4U26Sv3ytQXM8Bav39N3miiWpoOAniJwnn6JvgVLg0AYsiFR3PQZRpoLp7uyA2JYbrgf2xnsY7sMZUwMKW8X99Fj1/Ui1+XtuSZCG6CZSqMnnKFXGo0ZW+STLvRyEUryyTk7ZPV8iWHxj+drWayXuCZ5WZu4mANO6U7DZlkkhQDQPJo3nqjm4SuY93m2vjbbDub3oRePW5sCH8VLRd/tPAcLjW9CNLLPquTQ+SCvdXWPricA7sifbqFMrHVz5kZQU9dwqiKkYgXiqizSl6/BxPnU88m4ZvZRHrTxY2U4swg==',2,2,NULL,NULL,'bwb-integration@example.invalid');
COMMIT;
