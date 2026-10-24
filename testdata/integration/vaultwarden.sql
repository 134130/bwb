BEGIN TRANSACTION;
CREATE TABLE __diesel_schema_migrations (
       version VARCHAR(50) PRIMARY KEY NOT NULL,
       run_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO "__diesel_schema_migrations" VALUES('20180114171611','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180217205753','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180427155151','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180508161616','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180525232323','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180601112529','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180711181453','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180827172114','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180910111213','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20180919144557','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20181127152651','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20190526216651','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20191010083032','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20191117011009','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20200313205045','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20200409235005','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20200701214531','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20200802025025','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20201130224000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20201209173101','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20210311190243','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20210315163412','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20210430233251','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20210511205202','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20210701203140','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20210830193501','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20211024164321','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20220117234911','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20220302210038','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20220727110000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20221018170602','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230106151600','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230111205851','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230131222222','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230218125735','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230602200424','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230617200424','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230628133700','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230901170620','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230902212336','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230910133000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20230914133000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20231021221242','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20240112210182','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20240214140000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20240214170000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20240226170000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20240306170000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20240313','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20240605131359','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20240904091351','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20250109172300','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20250820120000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20260309005927','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20260425120000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20260505120000','2026-10-24 12:06:14');
INSERT INTO "__diesel_schema_migrations" VALUES('20260902120000','2026-10-24 12:06:14');
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
INSERT INTO "ciphers" VALUES('c26b4dea-cd47-47fa-a68b-0f83d040cb8c','2026-10-24 12:06:15.548328897','2026-10-24 12:06:15.548420480','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,1,'2.goi5pzRmA4DqutYNVc2l4w==|Wih3KZJ8kiCkfsIzBjL2NA==|Lb6or4PdzgV4TayBgh8ZmJRdemWfFMifxPAWlCwN/2o=',NULL,'[]','{"password":"2.X6admCCcnwdJzcjqTbwXCA==|RHwUFU+B41Rszk+aviwGniYhLtVjFZODffvXGfWRFs4=|7zbIL1mE/8Re/3e0O4BpQGpDIcnaMEyZAyZOL+mdB5s=","totp":null,"uris":[{"match":null,"uri":"2.vMBXlfzQJ7HuUtH5sdYHbw==|KJcICDjee6cgGjomTMg86jnuysaXiAkPPZCjHjnvLus=|/T81SH3rbziHOBwXMmyKF+5OT3xWUkqjMR43ifd1V3w="}],"username":"2.WdklXMtHH1FmC/Sw1k4cIA==|lLI7ydp2uozVHtXr4lg2BaUhbaMpXuwmQV/wv4su7ec=|0ov2HgukP3d3SVRltqO6o35PNmTXSDlkhUOXnozDhVg="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('ef2a5a26-80b7-4f20-be1e-4c744e559c8f','2026-10-24 12:06:15.550167396','2026-10-24 12:06:15.550228105','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,1,'2.vaDT/XmK3xjAngmzLpJ5PQ==|geth6zKOv3IbcZgrI8T47A==|05rUMHKOri7fTNFghKfQSvhP3Ac9F54YjOjRr8G12iQ=',NULL,'[{"name":"2.2ip6eRT70+PgI3EO1EV/gA==|vxwsToAc1aCs1qxIP45eoQ==|DZusTkqk0056gw+EddTcQgCtxq20JDwa/T8IlRfNzLg=","type":0,"value":"2.bIbjRsObc+hB5B7UMm0QqQ==|6CMIle3DQXJgN9GGurFCJw==|j7mWIqMr/Igo7FtxcuNLWS5+dpkrS9ZNtnPdE9ByLuU="},{"name":"2.IJ/w0hBi7YbLrUuqwUO03w==|wifY6CC1Bb7LJqwt1cW8Og==|Vke2AOQmYPLxG4fiFkSxjobjotfE0L9cVMHpzgwLC5A=","type":1,"value":"2.oW5RCIKkYJu7AF3iU45GQw==|RKuwbqklBKYo+gcexEohJQ==|AYVzMcdnxsQMiRCHh2YSlkuqxgUiKiJ34JITh3pwXgs="}]','{"password":"2.ckm1C2tinh1jH8t+0m9pug==|3L0+k7yAHkcgrsOVgsKyFg==|o5oTukiqpfjew5Uc6WNrBguXdmHJfFPAIjdoOhNO4PU=","totp":null,"uris":[],"username":"2.GQgVYep1N55Jn3YibUhyag==|LF5vJDV25HmCedu200lI5A==|cOhLzmna89N4CH0rUfMJ2uzQRvBlm5Z3yKhFser4jwc="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('69d3b24a-bf96-4a9b-853e-b5b19889197f','2026-10-24 12:06:15.551797188','2026-10-24 12:06:15.551887896','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,1,'2.EJrx/1bBR4SisnECyFMlgQ==|LbfmEa9r7HtjJvD9kOCabw==|zE2ORxfydRfr2Rg4noh72C6hRxJ50dkvgrFJMKOsdyQ=',NULL,'[{"name":"2.1eJzuRB3pfqapcbv17SpGg==|R/whwKSS5dsB76dYfKtgtA==|L+eXplLkBWojtneHRkKpFV7qwWiaukQyXmTcUFtxkg8=","type":0,"value":"2.K0rLxcYend3rvOINYBCMCA==|8vPaDu03qRK8wKJdQpwLGQ==|01gkAmp7XCcT3y3Ntzs5hUMQ0lMsVcNyg4136iKHQck="},{"name":"2.C+v08lKyhH57JSSgIXIl6w==|EPeyZGQPV8L2/OdshSHZwA==|DIVoyVhpg5zyp+8kRlisvWpKKAkRJ5djaeadNftcMH4=","type":1,"value":"2.VMQ53o3rQtfrl6HTKtXEYg==|REl6xnU70U2JM2NU2E8mwQ==|OwnHnSIl+U9QpawIW953swYq+uN2d6veSegir5UbD/0="}]','{"password":"2.FQfNFuuQr8sow59Wwt29yw==|V/D5PJPfzYPRGdCj9RoXbA==|tk56wd/oaEeIiJvn3PgpvTuGeiFAjoHfAzw/9W74SNI=","totp":null,"uris":[],"username":"2.c8eyx814AvKhr5ST+SLvsg==|GYW0i7I+Q00dz0QQSjbNaw==|TlMxhPZ5ku/8eGLqoATwVC1JegXhm4xaSSDrvHQeP+k="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('75af6c65-e8f9-402c-9376-af5099f881c7','2026-10-24 12:06:15.553390562','2026-10-24 12:06:15.553474062','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,1,'2.2cvw7q1rR1Rarj+RDi8ggQ==|Dx/qBpb9I39/ZMd/WbpGcTrvTZ7yi6PkNE9fWxwe3xs=|TdGGRAurroI0UrmZ/bZfhZ/BM75wwevw87S7mClXbYM=','2.//Be3Yl1Dkn+bABUMB1k9A==|xLaLt+f2I/zSJZoonAj4GIFV4Wyl2anM7DjXJLJbomE=|23LE2KDetxS/SjRGb0eW7MB2473auyf98i5vXrxsO9s=','[]','{"password":"2.5NE+pZ9dg19y75Qkwru+Wg==|NGeOtyEgOmLszjys2gVHuQ2Jy3YDi5WyEvSbbVuRSv4=|OE5lr+BrK84/+NGqa3HBsbCu+AfykVTX07n55JPsNfA=","totp":null,"uris":[{"match":null,"uri":"2.fC0kGUgfN0td7CSE9mqqgQ==|L4lnM9HPUAGU8Rh8ni3eRgwFZHUNA+8U+MsKrdTQfgg=|H8nTGKglh0SCfQh9kpUUVVRxcNlN52g9I2FI+EHTJjw="}],"username":"2.TJ5PaTLiFQJ5RxuDSep9hw==|yIruDymk5uMGYD10Yg3BlOUOn9Zc0WFheewTRX0GFEw=|WdsUSBGTY4hRJPNzOmhePV3Wy+WfsU6j9dFTekmCkPc="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('dbe9af6c-c830-4f0f-80b0-c3ee536da0e8','2026-10-24 12:06:15.554876187','2026-10-24 12:06:15.554924229','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,1,'2.CzAv4PCVZcnr3JvOhDqX3g==|Rv2xRPR2q8jUYHIqL21gUA==|OV2brx/a87ev/pvfjLCUBeOxNouHSm5uScSbp87Bs4U=',NULL,'[]','{"password":"2.X5RTOoSMdbQcNB51T3YiTQ==|HyvXJNBcv50zb7r9vcIxhg==|RIw+vJDV7LJ1LP5C7Ibojz9VeZ6lKG9Yq38OcUAt7M4=","totp":null,"uris":[],"username":"2.b0Yug3cYKSl5+zUEKiGTPg==|/NeyyaHtR0M6UAaSKzaomA==|BG11PC1fXO4gMPi2Fph/X1ezpcxa7qEY9336WR5iuCQ="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('b9359d77-c8ee-4d67-9557-f64d7e2493ac','2026-10-24 12:06:15.556326020','2026-10-24 12:06:15.556408729','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,2,'2.8uTxwoItg+WeBdUXADLAIg==|q2wPvh4OBS45JM2YksWP+CEd7FCu6dpRi8xQubzrfR0=|BgtLHLJYnoLkpa+0LLRXAzDEtk1VVuHndhbPjraJnjs=','2./gztqU1rUjplWXUTWbyAcA==|GEQG/HL5ApIJ9cLL2aWlIvzGEh+pYdqCcStymx5DQVA=|jGJiKeTkTFd+mFaB28Bp7we34+T2TyupKtonFCWf0k0=','[]','{"type":0}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('21cd98e1-8669-49fb-8518-f710856702f8','2026-10-24 12:06:15.558403145','2026-10-24 12:06:15.558484353','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,3,'2.1v/qnuiWWGEF0QfmhhV3/g==|rIkl9YPmLiU6b0ZZXgOe/7hiWHYXfY/5HLJaym5SY+w=|kpRxacBKdLum7B/6YhMoR/ccaaT/6ofK+HqcXYkc03k=',NULL,'[]','{"brand":"2.PqQusOhohSba+1kd9oPoqw==|W2bufyAwdw+DvE9qd0JcbA==|QNuqBYNTuSbcGgnttaFeKKQszBEWgnZ/qsCep1wV8kM=","cardholderName":"2.zVYbc0b+WAmfNmaYwmU3sA==|kJhkNZiL4aG0VTW+3cuoOg==|VuLvpYCyX53EFcfipXAJOYYJP62ORjtyCRGzl9JpAA4=","code":"2.R+NUbFQF5tOENkgGudBE1Q==|5o8QQQ8S5+xjhEbakjOyag==|UBDPBle1burSpIRydQgARdGMfe87UWQ1K4KxLPpkrBU=","expMonth":"2.slmVLVuQQXoBGFrvBRB6Sw==|oTQGcrbFrkMmT+Astys6Hg==|1w2pbNCFCorSvwfup0gM8DW2b4ijw3d6C/7hiyxtssA=","expYear":"2.4p3LXAZOLTFnKZGF9ljBqQ==|h5/kB0XtMGqSjF5awxm83Q==|/FmqNsVGUJukCNNZybcoB95sX0sE11WMcvPs4RvOBpQ=","number":"2.M6w6GSG7Irc+ephYHGl50Q==|nFqjF7QneoutFSCTWNvXmOW+JZO//brH15IDHSWNgu0=|JPPeL0ojVs8dfu7cvG/M/OWkhzWJAQqkFLALlZNx3ew="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('95d5787f-bc99-49f7-bca8-d7717c8c3150','2026-10-24 12:06:15.559990395','2026-10-24 12:06:15.560047353','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,4,'2.pe7v/0OXhiaijQvIb68C3w==|cfXbztrbXdbA8ONyDL5QpA==|7dd5h/bYQuyEHv8SeNWhzpFDYOHL+vrvGYf5Ug0F++E=',NULL,'[]','{"address1":"2.nYPG6uG+Yn0pkHyDlaWagQ==|Hx+Q/L3zKpYucPyNLl83mdBleCIGZkTatKSpwlMnaFg=|OwU/USrWeLBJSdefWxsnTz6U9nAWt3cqrFYLunElI3E=","address2":null,"address3":null,"city":"2.gkBy9JEznqamnig3KSLENw==|5QlzwuVTkBIpKfXpmx5Rmg==|FrEXRjsBQrPw9H+KHBnoXA3LNqWoRNJGvYg8uon+kjE=","company":"2.O5ysD9KK3ojeb6Nm+ttAAA==|6khE0TjYnyJw8Ql7LXJc3A==|Agwty2C7vf3JkMJIOPUYEaKPwxDScAXqH29pFfUBpU4=","country":"2./fP6lQeanGgmdez17HyykQ==|kkNdZNs/fbhY1KESpR7o5Q==|7horVJ9yKEiwZPLGb/NmysxqFaIzGqCgcwzWUuttRAA=","email":"2.NxqLMg2peh2OpzNuwOegYg==|9K0qVuiMDoL7YbNT3B0hdryXLhtbd0f/uL7Qw5543zo=|f6vvi6i2DbD5thVv/G+6W8AEE7o5GOR4X7YiNEDUQdA=","firstName":"2.YJCMyCLgGdGFReZ7lRnpXw==|N+d4UHYOSrViibdlOwMyng==|nDw7wClRlt91lEaKlgq7rmiK84kpuXzdytKuGswa174=","lastName":"2.HK+gm7iLtliPivzr9pHRKg==|qf8EBk1Lzwzx+kZe/45rIA==|fnS8JYjYTJ6dxw2ZT8E5H96RtpQhy+SQdxW/+OVv2DU=","licenseNumber":null,"middleName":null,"passportNumber":null,"phone":null,"postalCode":"2.RKNfgbEg6973Y7mCwz3YcA==|2lPnTQJcrT7xk6RH+M34ow==|jrtMutBBGyLxwJVGvwHTWLEh6E2Ubz+Z0RgjcKqmHi0=","ssn":null,"state":null,"title":null,"username":"2.3NulLvAA44ficheVi7sHaw==|vyFKDALuT94mpjany2j5SA==|DhZGMTr+WEHsZ6I7JLRvpy1X0bslKb3aKJUufDj7PuM="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('e2b80879-1975-4ec5-a263-3b820f6dedf8','2026-10-24 12:06:15.561383769','2026-10-24 12:06:15.561441144','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,1,'2.frOdMQp73lLYQoZHkGmK3Q==|R5niGuA/4MgdZhQZGDkYJ91PHoOQeFUnwrSZHvXd+EY=|Vh9jJGC/XXcT4DVvJOTJYxPoZWMD/ZpLes3X754aK/w=',NULL,'[]','{"password":"2.cf/JozBNCf/EB+a2bboB5g==|vjVAP9diUlAZIHd/+bxwIHfMgUMrtFCRr7yuJzvLnkE=|ZBXzkTGZ+FAvLC1CJFzAD02Dsa9Rs6RbSlzptwh1+Zo=","totp":"2.DBDCprsowJ55QGc0mh9ffg==|bnEoDMQjFEYXpU/zD/cNOpBiXbDhGPcl/Hu1FSk83EMrGEJm2ezo7ma9TXFjrytG|eHYAhlIUsQzl5ZXnVt8ttKU7KIPaW6JVKK/t9EctW+E=","uris":[],"username":"2.m4f4/eIC54TdnMjNvZ20Tg==|A/INZjQmRNszWzwQS7XoM/brbWra/s5mtXk1dE39MDA=|WK/CJDOPE1duRQzjWCMYfyGfDdJvdc7oeABy3MXeOr4="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('87939c37-7290-4add-a82f-93f62db07083','2026-10-24 12:06:15.562638144','2026-10-24 12:06:15.562704602','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,5,'2.te4BoAUyResYi19nVl+ARA==|9amZ3n36T3o0AbBeAZ+uFXEgnJZJCKmu1T7r8ryR27c=|bcHEJs4mOOyInTguoKnn+c5WA1Rg3NwuTxMk7SyZYXU=',NULL,'[]','{"keyFingerprint":"2.v4Ay3KqYhyTz9un8l23+Bg==|cJmrPrnoe5iFeC9SbhTbxIP4xm1mR8xikJSTxFb+FYisqZSFNs4F9UiHtPAQUUmAnEhfsXSuMA17sxIkb66kxQ==|WieS4W9dcCSHVMotdC0U19Qs6QF3LpHaZ9AWBPnIegw=","privateKey":"2.XwFkgOwCGlTDNiP0CCIdIQ==|TILuXdUK8UYnM9eXw9oBh3JNwqdBimJ6HjikjH5W2k9tY0eoeIpI2SazbC8HMKd0ALxbaSrzKzIvqtuC5Dvf40evVLHssdk3VvWbiXi8+DOvbVMdVZEevKdrPakYaS9/gMIlmt85Xuxc+ccx3rL9EqWJ2ZDNEvSQl8YxbNnTAwxgUJp8qpPHo2bQ/aWydXgJSqF7O+GfouOku++lGPA0uPUj8haVv6R90jTHso/4qy4TjtT+6shDUsuri3JX+9yhg2cDMlCYOXq8AKKVEfM1Y5oeBRipTh51OWTvcl9VpUR2Rgo1DTMPsUu1Dry4uvphC954Bgmj31T5bjvfs5/70uZps40RvWfu+ZcSsTeWNXOlfBoRkQ+t1HHB/4W6/BSexkWmm3f3BxQN50a5DayY1rBlEPVD2imR9oZvd8etPSUDUh1lwVXiQ7TM1D4wvSKkhfwfdIcCs5tJbDV30c0Vlih3WlEWqerAsyMIvIM4gDzGcpc3CkpkvfGNORjxNzhIoPbQZG/ozxJ4rX5BOJTnDw==|5AjsWnP71oo/EPgDRtgX3btz9IDu9rs6bXa2ptJMRw4=","publicKey":"2.dGmoaiSkdGBb+3TRk6eitA==|pOWViGmb2oMfHlwMR9Uu05Gf/xtAXSYO27G82xBm9ZqV0loeOFk5CSyiODzvPVYhibmiMNV3pJr5YD8vlorCcnAXd98kbifFx+Q8rAn39AKKI0lsjiBSEN2fDF208Jbb|cRsHKzbN3hgSMgQbzsxeopZcA5U5Wgk25AaUpqboWy4="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('20877946-dbf8-45b1-aafc-b09b7f54fb1f','2026-10-24 12:06:15.564098185','2026-10-24 12:06:15.564151894','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,6,'2.WBO3Zo8gkyXaS3GxB3KEsQ==|ZfZddMWmhIsDdt949g5gFoS/VwzgUbid+F4BmCXuDwY=|Ss9E7gtsiTuFwgBgLoR8om36KKnocRcQjz3BPUYa65A=',NULL,'[]','{"accountNumber":"2.WTbbDuMLNi6SWdGVXpugug==|pWomnKyNuaZWRdU3OcifuA==|UVYkDwlidNpj2lRF+bN42OU8a73JpfTt65xkR+hxTek=","accountType":"2.ZxclmnlW4ilMI+akH27fiw==|e/pj8+6fW17AM9SXF6AU8w==|Prx5heJVQFEvEFKt+m15bhEky+uJ1ROiAGx7HC/1AfQ=","bankContactPhone":"2.55hisLhbX8KZoGIOyphdjA==|dkTJj4mxYQUr0hkc4lIRhA==|cNSIABKmZuinHwCjDF4CC4nso7gsZ2Y2jzSZBRqkFUY=","bankName":"2.Oq0wqwk0Xm/HKiPHuLaVBw==|Ch2e3uF4mXcEq4UZnMDIVg==|USvvM1AmGprF5VXfCINt8v+23lpYNOJC+bmLp99IPcs=","branchNumber":"2.V0JeTONoSG5mQAuI6xqOxw==|E0joY4roK3DJYWUBQ58izA==|m3F5JtCk2jIHRYw2A0CG3zprkaL6VDjBz5S4x3EjYiA=","iban":"2.Iiq5alv5MjQ2W3e6D7XTBA==|eL7OP+Fwx96nXnHiQjRRKg==|ZTCVNg/t7refetZt+LHf4FcO9J/BPemXKXjpMF0ZU7k=","nameOnAccount":"2.12X1pfR+ae7YWOMgy1IXOw==|rGZfqjGqAXucR0Q1pGonlg==|P+Uq38okS7G54G0vq2itDNHPvqO0SewanbzOr49llnI=","pin":"2.5INTn2v1kaLE5kX7m1R7lA==|y+YP4t58VYqnBaL9vOsZIw==|QcHByVFQOxKbmTJEvMdIsw5EDj4UplK3PGChNENw/LY=","routingNumber":"2.jHSS9UQKtaR9NG5K7AasCw==|1OuR0tFBp7ssBSVOHktPNw==|bX3W/JqbfVvxYPVIi1eZNazUjFUEbs0CEuLy2EkLK9Y=","swiftCode":"2.JX0x5Pt9rCYvpih6zhoSlw==|rSxjyYdmn6Bk8IKPyCMpGQ==|RECU10rJg5pYT2VbM4PHlMt19e0F5KIrTBaf44o3mbI="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('0e76dd24-3634-4560-86a9-f0f148d7110f','2026-10-24 12:06:15.565389560','2026-10-24 12:06:15.565431644','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,7,'2.qehZDA+Sc/kl8At6ozX1YA==|nVOc+57iMnKoGwXbKIR6YXxMT9afVVABSIRvpSLq5YI=|v40pkntM99zircasC4bZp2xbBLOFStmL93kD6lMzp9M=',NULL,'[]','{"dateOfBirth":"2.eHWmDh3qhx5XzcWRI9Gl3g==|NNpHOJigM2uFvG5UjZNb8w==|xaJsFBzMEh4ZJWATbFjk8gicnEr4+1amDfyAxWl9/dg=","expirationDate":"2.IV+w/ZrfI2YVCHxCYH8RcQ==|gQP5+gL/DC9cml0PG+wizA==|bSueS3PlEM/IWp+6BhFS7zIokNYm4Wv5PLiprsJka1g=","firstName":"2.T+/kAib291Y4H9BDzcNIlw==|9FWeWhO5ewxHGmY/UtkgwQ==|exV/tjvAbsYfhhofs2M3zYRyYu3qVeCeBWvD08KRf88=","issueDate":"2./NBD5rDyQSYnHCh2qxkTlw==|sRngLkc7xvdNRpQ4dGCB7Q==|xFuoPE8sXoikNId8hF/V5yH5cUokSZEupAtRE70rsi8=","issuingAuthority":"2.XgYAQGtWOK0bQQpOWcQp/A==|XG+9sQQE1sIVUao52f/XhA==|9Np+yB+HxpJamxQ4YxBmnrmX+7xhUcSKGIV7dpTU8hQ=","issuingCountry":"2.WvypvDsU3FaiqG9sitdJNg==|62qC2knO42W1YR47lhWN+w==|EMorjw7nJyiuJoDtc1TqyBKOg+sxhuEpwwPsSuoDRXQ=","issuingState":"2.3hjplXPS2dZGYp5YHnhCPA==|eymfWuOTODOWeJb2LO5BNQ==|8tE1YKz+PSANOd8KeTNPBrisWDiLJ/G0rVrzxkR36IY=","lastName":"2.HO9+EH18F+QXNQCFEu0iMQ==|1kRyIbKQRaL3faRXfdwzxQ==|oeDlDtl5q7yAc6QgL+SpPJJjp/R2M0cAnQy7GMJ55tc=","licenseClass":"2.q3KFPwArECjMzQ4HBA6pIw==|Bjhhryv3etYT5D3PSqDSxw==|WNfOcc0yLsnmYH/0n+rNURv2B9zJKh68sZLMH8G+43s=","licenseNumber":"2.VdT6xoNRs7Q348XUzxQpug==|H7MoA7tY0OhpSeD42CABEg==|qo2BLPP9Ky94gFdrVnyIVUy6QXGW5RK3rVW7xaKVw2Q=","middleName":"2.2KCx7rXVUEOE9LC/BO34wA==|gy+DWD2L8Aw2hMi5+avKbQ==|Q9mO9QHxee12ciI13BcdrH3moJABtNOINMcNhEntLCM="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('60a0fea0-6381-4c22-bc1f-c9ca7e1c72e3','2026-10-24 12:06:15.566903268','2026-10-24 12:06:15.566951977','62bb2087-2d4d-40c8-a488-60c65aac26a6',NULL,8,'2.Hif9xFzJDHjnTfnbpPc/rg==|HhMOVjnl8OGnTNpxIi2H6w==|o35zz7qX0A8ClBnASySDRzzder7mTkvHFOAb/p6EhFU=',NULL,'[]','{"birthPlace":"2.OomsoIMHavhjcTdPyInE8w==|09B/TjyfuvVPFZyS8S6mpA==|3XQIhaFajJZIIX6eZrO/PmQjP3vLOYhqCKTUoisDyNs=","dateOfBirth":"2.gpy1yDGDXoffNTtH2bpO5g==|ady52QBUHFX26htYR2aQ1Q==|20maMRzE4/cB0TAbjUOpg/5KU6T5B3zrStwx+M2b8Yc=","expirationDate":"2.9BYHJdhMxQLSa2p//Yh43w==|n0IKzcg1/x/hvdxlKCu3OA==|7vcKCTQpgET7AQOPGNyiYrIXC5RHtBzQbN+s9F6PK6k=","givenName":"2.AGR37u0f5bqR9mgzs4WbnQ==|GPp7mXbF0PRCsCzSvrth/g==|ld/EtReX5kx+9LO/FQA8oEWoBk9wQViPVRznWLSAPVY=","issueDate":"2.ztFlX8UxqcNeFTYoGuqyMw==|UBBn9HbRiwfK/Ho+7uce3A==|LPiYnTfUKHGJD+dt/RfXRGPagtFG056WTDWo14q7BlM=","issuingAuthority":"2.mjc1q1q44FxP2tx+YC3Xsg==|Bj4Ov5sWDgHJwrd1Mq2O0Q==|pmlLDmBYs5Em363CIRqi3MiDgfyFMNGRCHahXOsTrhk=","issuingCountry":"2.WQPcyU0XcX0zrGuACSfrdw==|NLwZov3cTxvJ8dQoMPVr8g==|rmpDsT6bWpVxQ9IXUwj5wecAoIbKE/4a2nrbqTO9jLw=","nationalIdentificationNumber":"2.1coD3nOuqcjOKkl0iun7Rg==|oVoF0vFBrBeiUCr6okcKJQ==|308WxrVR8jvfDXmI+bcyMOvU95ky/bvfTk/Kn7QzhRQ=","nationality":"2.GUweHm6KZB2pZDH3qxY0OQ==|M3vVnJu+4GSQzGATIhaH6A==|tnkbIeCK9NDT0fLLlS/mb/N/1cIHb5oS0mzfu84Ln7w=","passportNumber":"2.RoWXUtr4vFbLyIrTDWk9HQ==|8H82bSWFD0AVUjvaOCxzfg==|7ntIpQtyykzsIn93gOcmktZDjaWYis4TdHEVbblPokg=","passportType":"2.NkfuS0qMFyJSuiazPatsQA==|Vz4fHvVMYcGM6vD2E+OuXA==|p2KgCG+bmBiWVzrvkAEg0C9+sFQAZt5UulqxSJVhmYQ=","sex":"2./ll7Ye+1wxRIfhpi5QryQQ==|JnGsB39Yl8vzSvUbAs7x6g==|nGRn7Q9raebycUkobX1Vs6SgxGr1XmfW+Gf7poqZl2s=","surname":"2.b5z169MoViGMP+otk2Do+g==|GO7seUDdPrH6MTzSCNihIg==|EYz9DjWem/F6NiXy8fWtd1FrFi4f3mYMWbBwqQofmXI="}',NULL,NULL,0,NULL);
CREATE TABLE ciphers_collections (
  cipher_uuid       TEXT NOT NULL REFERENCES ciphers (uuid),
  collection_uuid TEXT NOT NULL REFERENCES collections (uuid),
  PRIMARY KEY (cipher_uuid, collection_uuid)
);
CREATE TABLE collections (
  uuid     TEXT NOT NULL PRIMARY KEY,
  org_uuid TEXT NOT NULL REFERENCES organizations (uuid),
  name     TEXT NOT NULL
, external_id TEXT);
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
INSERT INTO "devices" VALUES('30000000-0000-4000-8000-000000000001','2026-10-24 12:06:15.539287940','2026-10-24 12:06:15.859857424','62bb2087-2d4d-40c8-a488-60c65aac26a6','bwb-integration',8,NULL,'ilGY8tENcCMNCVmm2aP_enUbjEQ7kX3WubD1HYd_dj51XjZu9M3H8G4UkyoC1Wgtsr89cnct8Ai2UQStiEApvg==',NULL,'f6d2c625-69a0-4ef3-bd45-eccfe80bdd2d');
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
INSERT INTO "favorites" VALUES('62bb2087-2d4d-40c8-a488-60c65aac26a6','c26b4dea-cd47-47fa-a68b-0f83d040cb8c');
CREATE TABLE folders (
  uuid       TEXT     NOT NULL PRIMARY KEY,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  user_uuid  TEXT     NOT NULL REFERENCES users (uuid),
  name       TEXT     NOT NULL
);
INSERT INTO "folders" VALUES('11eae1e4-eb45-4bf7-9019-4eabd12919a6','2026-10-24 12:06:15.545549314','2026-10-24 12:06:15.545617355','62bb2087-2d4d-40c8-a488-60c65aac26a6','2.2P3liNxV36BK2TA9PnEmUg==|ScUrSPsgCvAyMIhocdxAcQ==|uaVtsXlbJs4KkmRRhoWvaVYIrCUW9HCxQpkLCToDAxY=');
INSERT INTO "folders" VALUES('be0d2006-099a-4360-8e3b-b93b0971322a','2026-10-24 12:06:15.546984689','2026-10-24 12:06:15.547035480','62bb2087-2d4d-40c8-a488-60c65aac26a6','2.B8T6fgmJbUBODEoIDvYbTw==|00BgdvNH7NeQoDll4Rj2IQ==|CNZAP1b1WJJDVSI1c44O2FtL+dDKR4JHkr3TZv6hTs0=');
CREATE TABLE folders_ciphers (
  cipher_uuid TEXT NOT NULL REFERENCES ciphers (uuid),
  folder_uuid TEXT NOT NULL REFERENCES folders (uuid),

  PRIMARY KEY (cipher_uuid, folder_uuid)
);
INSERT INTO "folders_ciphers" VALUES('c26b4dea-cd47-47fa-a68b-0f83d040cb8c','11eae1e4-eb45-4bf7-9019-4eabd12919a6');
INSERT INTO "folders_ciphers" VALUES('ef2a5a26-80b7-4f20-be1e-4c744e559c8f','11eae1e4-eb45-4bf7-9019-4eabd12919a6');
INSERT INTO "folders_ciphers" VALUES('69d3b24a-bf96-4a9b-853e-b5b19889197f','be0d2006-099a-4360-8e3b-b93b0971322a');
INSERT INTO "folders_ciphers" VALUES('75af6c65-e8f9-402c-9376-af5099f881c7','11eae1e4-eb45-4bf7-9019-4eabd12919a6');
INSERT INTO "folders_ciphers" VALUES('b9359d77-c8ee-4d67-9557-f64d7e2493ac','11eae1e4-eb45-4bf7-9019-4eabd12919a6');
INSERT INTO "folders_ciphers" VALUES('21cd98e1-8669-49fb-8518-f710856702f8','11eae1e4-eb45-4bf7-9019-4eabd12919a6');
INSERT INTO "folders_ciphers" VALUES('95d5787f-bc99-49f7-bca8-d7717c8c3150','11eae1e4-eb45-4bf7-9019-4eabd12919a6');
INSERT INTO "folders_ciphers" VALUES('e2b80879-1975-4ec5-a263-3b820f6dedf8','11eae1e4-eb45-4bf7-9019-4eabd12919a6');
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
INSERT INTO "users" VALUES('62bb2087-2d4d-40c8-a488-60c65aac26a6','2026-10-24 12:06:15.302311941','2026-10-24 12:06:15.567050727','bwb-integration@example.invalid','BWB integration',X'C6D40D9BB649FD29B364EE9DFA650606F5F2843B4FD9D26E5C657B4C3531739E',X'21BEFE93900ED3B62CA69B3A2EB5B12A4A7A307FD0CA0A0748FC121C66AD883DF33ECB8B5459953C21D0DB090BEA9F88EE0F70CED61400D40AC8AB6FC107824D',600000,NULL,'2.HWdh3tK7iH4IMoG+6BKT4g==|q/eU6tcAOu62srNJTKrZ6MCDLKYaNTuMdeKFV85ssUbARuF0x8bpVKzuKfA9busEIYNr/wL8jN80FkAj+E9EQ1WNLT+iU4sn8CdJ8p2mkNw=|JGOAbDcxzIra+ehZWXhlWpsuREFKL81TQudGA5R+5EA=',NULL,NULL,NULL,NULL,'c5a88488-32b8-41ad-82df-3fcb2fe02d75','[]','[]',0,600000,NULL,NULL,0,NULL,NULL,1,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
CREATE TABLE users_collections (
  user_uuid       TEXT NOT NULL REFERENCES users (uuid),
  collection_uuid TEXT NOT NULL REFERENCES collections (uuid), read_only BOOLEAN NOT NULL DEFAULT 0, hide_passwords BOOLEAN NOT NULL DEFAULT 0, manage BOOLEAN NOT NULL DEFAULT 0,
  PRIMARY KEY (user_uuid, collection_uuid)
);
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
COMMIT;
