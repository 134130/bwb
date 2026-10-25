BEGIN TRANSACTION;
CREATE TABLE __diesel_schema_migrations (
       version VARCHAR(50) PRIMARY KEY NOT NULL,
       run_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO "__diesel_schema_migrations" VALUES('20180114171611','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180217205753','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180427155151','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180508161616','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180525232323','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180601112529','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180711181453','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180827172114','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180910111213','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20180919144557','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20181127152651','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20190526216651','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20191010083032','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20191117011009','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20200313205045','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20200409235005','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20200701214531','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20200802025025','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20201130224000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20201209173101','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20210311190243','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20210315163412','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20210430233251','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20210511205202','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20210701203140','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20210830193501','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20211024164321','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20220117234911','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20220302210038','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20220727110000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20221018170602','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230106151600','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230111205851','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230131222222','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230218125735','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230602200424','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230617200424','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230628133700','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230901170620','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230902212336','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230910133000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20230914133000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20231021221242','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20240112210182','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20240214140000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20240214170000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20240226170000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20240306170000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20240313','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20240605131359','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20240904091351','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20250109172300','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20250820120000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20260309005927','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20260425120000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20260505120000','2026-10-25 05:59:03');
INSERT INTO "__diesel_schema_migrations" VALUES('20260902120000','2026-10-25 05:59:03');
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
INSERT INTO "ciphers" VALUES('306875da-d69b-4625-b177-fea2df37f351','2026-10-25 05:59:03.818678503','2026-10-25 05:59:03.818772211','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,1,'2.VVDZAYbqFn5XTGmFbIHGbg==|W6HeyA6NEFyj8sNIsXxWNw==|f2BemhAdP+m/B0sUFzmSNSgMuaP8QaiXF28KR4azZkI=',NULL,'[]','{"password":"2.5k81d7S+D+60gOjjbhnmlg==|oq/BK9evIWzC1SyFfpMaC7/LrWAfLBPsAvKIfbAcv+w=|EU213WcrqcwcgiTN601Z2dLO3PTGH/SanBmkXUjW+ec=","totp":null,"uris":[{"match":null,"uri":"2.WZsjuwZ89GNKAXPtiWhIXA==|PtrrBF2RWc2TLOcEK+yD1cHoWDjD3CXNG4T2nrQXJew=|8/PlLt5I3okfn/bjRwg4YCZMnLtg3x4Ucvb256ZxYu4="}],"username":"2.izZu77UzE0MBcm6a3RVVhQ==|ylZsfpmqwP1qO3GZrNgcxuLWaydYo0TqAxqUsM93EWw=|dmVehQNdoLhS4sapmQLn648MOI0RCrQROfa2manFdFs="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('37296b7b-354e-4898-a4d6-57fb085fe9b8','2026-10-25 05:59:03.820182502','2026-10-25 05:59:03.820252127','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,1,'2.0Bqbln2PDETa8Io4SVYVhA==|ZgVjXOdN4dkiA0bEtUSWNw==|c42nX7c1U4NeADHj5nR3PK58eQQb9Qxj74XyImOCVEA=',NULL,'[{"name":"2.fymJwGA/m1KVrvX7FR3QJw==|TrEcrQfMXsGbJ6XFD2HxEw==|mgys5dr/rHIhNUB6AWfz7mjpf6cWZr+RbA1S+j4Hxt8=","type":0,"value":"2./1EDTVIgM79u4WaR31aE5w==|y/VdYzR+kbI7To17Zdn4xw==|93S1LfKMP9P4tNRcUEQF4V9xWu3U9PK6dTzW9thv+fc="},{"name":"2.0Im8ji86Zo71Ec4LUHVnVA==|jGebTEespi/gw3HV5mgMJg==|vAPCtHsSyffUguEXSDcS9QdfDvcPoe/p/X6k5ogmHa8=","type":1,"value":"2.4CbQRTiaUH8+JQ1iMDf8yA==|FIf6m8oG41ikShLMvfgECQ==|TMoE1SN45oi8/+tZykIzdb+3gAN9TJYn985lmS9hlMw="}]','{"password":"2.uQVLHJ4mFhTw9ACgIh3TkA==|gUEbMaVCCZyKHLeBIcxKUg==|EdaqWSmYAJ5LXgJkwGUmYxkXk9ZjLh1ls/TmWcHXpBA=","totp":null,"uris":[],"username":"2.NYrA5Y1WCNaDKYmkY8Q6QA==|8UJY/Mqba/YYBmgMH7M+8A==|9iAktoyEWv6KcSF0q9VmMzjXj6wgsBxhThNeMh3woxw="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('aee95654-6917-45ec-b1f6-cd4f853701fd','2026-10-25 05:59:03.821962043','2026-10-25 05:59:03.822109793','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,1,'2.XFmt2VGhkIyOyF55Pu6XZA==|JDIITBLSf/dnSL2KWOtCQA==|TGiTQNvBxmc7sjfvaCpRfECsIz9rWTWVLdblqVPmK+M=',NULL,'[{"name":"2.uepadhlAD+S4i2AB2fME4A==|wghQYKQtjnWoiV6Cr0Pj5g==|3xsraDDCwUD3AzZ7rQQM/jsz/Mak3Tm1XKhTWxewaFI=","type":0,"value":"2.toCinq+1TzrzX7KIFVb2QA==|ZZUYgh+mkl45iK2eUrIeWQ==|TD3kl2DkTT/5PZfOjXt40OaUCQ3wSfRU20MHI+8Bmac="},{"name":"2./OhnGtkEcKCAtyv53Cs8pA==|vj+/elZeBCQkzMwGRNfXpg==|Kdv24IWOXj9HGsb+Z71wLGGfzA/cFGLW8AS6BfYND0Q=","type":1,"value":"2.fdxXhqg5Pz2P8nQx2eSiVw==|Jkc7EoX8adsOxAMwYqcQVg==|trckTUFVxed5DkEOkp+xkgWzpTLGSCFnir7S2AoZoOY="}]','{"password":"2.xVTKTaT56zY5vHmSZvivAg==|440qk96worAOV6KfGq0F3g==|MexPLNNr7axeUpVvqbZVjXMu2fujJjcqonHp6hjyIQw=","totp":null,"uris":[],"username":"2.6ncvX42PGjhGmkdgKhuaAw==|4RDDA81pAWjh8cI9wiHNwQ==|rPICPfrwTC61bH9MX5u3v7pXbDYCRVYAEP6CdT4XRbw="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('77d96bf1-9b85-47e8-baf9-c9ea01604c50','2026-10-25 05:59:03.823311751','2026-10-25 05:59:03.823365043','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,1,'2.y1eNVWxMlAHCJuX9qj/PTg==|E4a+92JvwyBGXwhTsbJizeK/9RPL69O3295I6fpGxvM=|kTR69I9OQ4OoNTCY12PbjYE0gLLesAtvhL9urxjezrw=','2.+Xn1tSQfVVQs5+6XHwa+iw==|UK1XgccXL89TpBqxyjvUUZXcPB/0ERH2cBs04jZ2yNQ=|v7V3Rizs7OJn/ZJ3ORQgB6gdHeTbJcVdvqPEInDKq28=','[]','{"password":"2.nqNYFQcy/xQys5nDnLW0Cw==|p/HXCfGVaXLrlvrskN4LFj91y8ST4mcW5PGmk/moUWU=|Je+Q/dXBaPwRmljSwvjoddF5jej07snlKiZrFO375Mg=","totp":null,"uris":[{"match":null,"uri":"2.n2I6hyFgi641TYZRKMhZIw==|1nBwZzmLamkx3Ll7OKShi+6Li55Rq5xy0N+hC/J2ecY=|xbd/of5bu+HcgKlnLiEswKlqXy/Gwas0D24WURr2O0I="}],"username":"2.wucsF+yY5GS02oKTt9RlAQ==|FffLnt8mUAOP1YPYPLf0ymIpn0tc750rB5U/8hvbRG4=|xPPuPj1LZCgRr3D0Dt3aXTakkkL4uTYlHLmjrtAFrrg="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('5bbd4a4a-7d07-4bc0-a19d-529a57e3948b','2026-10-25 05:59:03.824442001','2026-10-25 05:59:03.824514626','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,1,'2.l7td9cqV1yxPn3ANbBPiTA==|1abXy14OkOCZH4ObCz+Uig==|333/MIdYsSR5rzJ2UTMcgYvAhDzrmlhpgNdtZldSJG0=',NULL,'[]','{"password":"2.BB9secgnxb5SK+tLXm/AQg==|gliePEm776rZbcA9hX41ow==|m1eOsTM6qmUyHJEl2pluFvPe7+tZc/E7dGMvmQwv6eM=","totp":null,"uris":[],"username":"2.bhTUOE2SktwnON749c7URg==|CblCMwC5NhAaX51yEIWYtw==|K1sKtbouJQbcG2OTl1Q4gVRc0OJUWe6ZPQCk23oHvnA="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('fa63206b-ef23-4152-a2a2-b95ecc04a047','2026-10-25 05:59:03.825655250','2026-10-25 05:59:03.825750084','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,2,'2.GexvC12uf2DY/qmV0+eC2g==|5jasxS/y5901sSODTx32kG/XsFVUvEJBGzEAwLcxc9I=|y9w5ekZAtuZ+LsMb8YueKbtyZ7L6skyymCwkIDMuhWs=','2.tIn0B/wC+6XTFQpINTQxAQ==|wNXFl8Lks1UjbtbItOv+nrhNhbg7CCqo+B4d9E/QV4A=|17+nVXmf0IEg9KE9dzyAlVAE2qgD0/7IgOp/m2t5E0E=','[]','{"type":0}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('61368e44-4751-4cc0-917f-e78fb7aba727','2026-10-25 05:59:03.827184333','2026-10-25 05:59:03.827272583','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,3,'2.9fPg6R6ySU4Vl/5TVgmIkw==|pHRo3O80cJeT8d8aVkBsxnIQ7W3pV+8HJMINAJZGrWM=|ZxOKz8BmRxlD/kjGEyDaW3g9UveKw08qOoiizOTiskU=',NULL,'[]','{"brand":"2./J5Fi8A9Rsyf/OkBaxIvfQ==|ylIrxkRRO21MmjENmdZ8DQ==|mE/lmZPU6VDNvjrW0Rmt6xNDY465U+KDdJn+zf+KLq0=","cardholderName":"2.1w2umzc7rJSnXMTA0X4jMQ==|mFUcB1ceYav8TOGu74k1HA==|K7blodXrvEUHXKpI09hox5PcoiPxD4s0SPDXJ2m4B+Q=","code":"2.xnc+/3s6tvQtzUzWnyfG8g==|62GikpTJX4424tT/NC3ScQ==|t54DSF8Rvd8wPVIIDw/RoM7sLAMKjhqQq+D67/Shcps=","expMonth":"2.Q/WWcjEEkDaIWy9MWMIEAA==|E1Rr5/bqX63We7Lz97kHgg==|IA+cW2JodtC9x8pEvMvTjwvn/VsuSmRMCtgitFVErQs=","expYear":"2.bhVbRYb2N10y3tx/XfeWTg==|qfU6cLicbYN1OHnmmD31JQ==|daNdfHqLR6FCd1o2tdyZLkATZUDuHH4xkbjNciW8oYE=","number":"2.VwwK3V/BAMJ3m4bg2P5LEw==|xtzfDeBnQGo7lpAi5YayVk4u9DHRd1JpbLOReXd0LIc=|omu9KdoUzSNTanyRd+Qh4PQMaAJRqQ/fuVng0Fqn9B8="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('a390138f-dc99-4969-a1f5-ea9b8da03a5b','2026-10-25 05:59:03.828573124','2026-10-25 05:59:03.828655374','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,4,'2.XOx0D/cAqP2tK49kPsfmAw==|XnUSXcGJBcBl9zf06Utd6A==|P20GYj5EP+f0mKWFfFzecwmiT0nLUWPJYqEaZaFuwWY=',NULL,'[]','{"address1":"2.m5Rt/4uN7nUtDxe80a/HQg==|zTmeXgjr85nmjARhCOJvlKccPSSMKJucot9Nt5QSfCQ=|TpJGBsWvM23sy5VRXeQ9ORDoOd3u1Ee5SnUTw+UPfwc=","address2":null,"address3":null,"city":"2.nRXpeR8psC0AjhXpySUADw==|RlQ81d9aTzlvbWaUekD2EA==|CiziG+SjVPx3jDL/V0gpGmPaqUykZIWymsyXIsS7Sp0=","company":"2.mrPrxrAxDgGyjoxtf5g6Eg==|n+OS5goLC4yFZP6qlHoLsQ==|CP8HYsmXuHdY4GUfEpMvbzdbuAQCl836VlHd6R1h2Pc=","country":"2.AyPtWvnAA7uiKuRHLf/f/g==|lAQ7QKcUZQO8gZq9aEKJTg==|lCsGCMBhKLSJ6o90CPYq3m9TLkc3R4zMi/uOD/nGRcM=","email":"2.ZfpL0oYB/ey1ygQNfu/bHg==|4f/uSC9lKny0Nc+3yGBi5T5jPRbGT+brghWfhxX502g=|BMHc0sxZyDTvd8ljs/q64Jg3ex5Oqsin+Et0Z+BtrrE=","firstName":"2.Yf3CMFudkRG0g6uarFw0QA==|j7R4aYoStsFSwzgEASYuoQ==|cKVXTRenbHm4BeGsEUgQ1aQO9PVkegd6ca01NzqItXo=","lastName":"2.wbgIY/NGfzlVGjI8LXtzFw==|LVeUDcJw+PSgHEVyFEHUVA==|2CQmgmObs3rMP1E/hf6jYgXoOvo7yx/w+5Fq+xawlb8=","licenseNumber":null,"middleName":null,"passportNumber":null,"phone":null,"postalCode":"2.c4TErklqm/LCUSDUgN4RJw==|csjCKUIefDfyOYNrwSwkLw==|Vo6v374quwiI5bNHOL4RDjZgFZbqtk45el63pHuOga8=","ssn":null,"state":null,"title":null,"username":"2.YEYLiqGYZce0H+VhVd6Ryg==|EHQToNmWvAU8fAYYp1NHaA==|Hc0eT389Tg9cskofNZrfePwvQRr+qi6/+Z6dcdGBxNk="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('20160329-5a43-4050-8080-74b54d13f21b','2026-10-25 05:59:03.829735624','2026-10-25 05:59:03.829813832','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,1,'2.Ko9g5BhS3AdxdFS/pa6NaQ==|WA4KlqjxEJP87OYKtXDDq8TrWK9sYDB5hU7MUaoH9so=|tjopTYTtK126KVl0mI2efP9EgFNq4ULGhX6gjEqULJM=',NULL,'[]','{"password":"2.7VHHleC2DUZQJFGA1ROFew==|Ti8jAsxI/CKRD4YLA+MmCLLamu/jQ8zRgQowPy89tnY=|cy4gQ19SoHp31KSoY0q1+Nq23pHuhDbEI09k29BHVMo=","totp":"2.L1EDbHvzujTocvT/pyp+Kg==|RHMAt8893xD8PVM4Nwoy8JnxVopg76nEb0ej8b1/Y3ddWPavjrI37T5BLFR1YWFk|AHzpAKaTTRtHA/AI4C9HmPzFblD59g06+CXQh+mr+hg=","uris":[],"username":"2.Ck+vHDlH690XNGfpk8HSKw==|ZToSfSx4wUXoDX7NPfbS5k8yN3gaYIonj5vrj4aWkNA=|OZW+7Ul+AN8LvMXdiLYqC8E7pnEMWWMP0Qb2cRRvam4="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('4fdfcc12-4d55-4682-834c-053753371910','2026-10-25 05:59:03.830892665','2026-10-25 05:59:03.830932665','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,5,'2.MK/TNLLluHe1lgH/YQAiTA==|QR2Ah9x9TaJ/3WEnVv3tBF2nhGf2hS8BJvlAelXtjR0=|KEBFLVMo49mFLLH68mteWc+QGNA8NF9NLO/drq4Q75U=',NULL,'[]','{"keyFingerprint":"2.DdTZY7UdJ2Ja0QVJUG83yg==|L5Hck0EqjPXYxW0fvbYMXM3slHCwosxv+W1HkVGgxr2DWD1jZyWbuXQRRF/QXdocP2js/E7k8o4D9TEnvBJ0nw==|BiZTDf6yoHe3E5IfaOHPbRl9OQkZ3nA9La4Acfq8WDQ=","privateKey":"2.ltCphEasE7EQTMGlxDcoHA==|K8Pho/qwWx0ewyaA3Hzn2w5uNVXfG0+NbEAWZ60YSTB3Yeq37o3lFXeuF2U98DeSjft5W2f7q8C3SSlhcTEEpZnn476n7U2LePfNU4StCwIfWdTbL6RXXxpL1/Xnpm1EmuUB05IhkSbOsQwUMiZiURvm5ATSr4NWMgXRT+S1VKCtKJDyTj1GHDgvp/sIqGKFUPP2QMKrsP6y3JRHiVjCkGoE0nVYKcTaBjm1Mryj5JeC4hMBeokMmvjMMcuII4mKA9MYBq9WxW94107YxYpZyARtNPABgEc1xAICZ1r8pru+QLU02OJaTxi5o/MfmedM6mvKJUIEAZJ7itD4qzHWvjcrQisdiLGIztjMqQbJ99RYQa9SU40yDRG3meKz06Fi5WgkncXFo1qqplJmvnol+WKwU/letuI/xjtA2VcQQ1q2EgCeliIeRwoeoU2uwfEab1k+xZ7lJioG8DKW7fBLbOi3srOfZWVUsmlSSvAszOSzK569bVZwiwdfNSCxzXy7L4Gto8uPgUAN9wpsOuuR1A==|7HLqsBelNYXG3dudNPDVCalUIvdOXlKtLpjzB/RdpGs=","publicKey":"2.7l2AXTF7WW9OCr4OE+gtfw==|uZ+u/S1ka1RzknqSykPpJpLJ0amlxESnCs9TdwDxoESjEpDqvKImYKY8FZyENN4h3YVYWlA1p3XGsVZSlReM1rxO9wcybIaZJK977dVXGNQDJrsFZWm4Rk+RN5zRAlpt|PMUGX1p8LgnwFdVfoUNq0mtmpvvuzYp4dLkAMQx3ed8="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('c5dee506-4868-41e2-9d59-0cd33fdef665','2026-10-25 05:59:03.831998581','2026-10-25 05:59:03.832056165','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,6,'2.KBxlucu6zvPLtY9AXYLlAg==|0ETbmoPifVYrpQhtLF1kvQvqrh0XBmBE37FturTcpU0=|v5uWPX8jQqt45CXruYimih1EmoVNPXLHSTRHGLBRaFs=',NULL,'[]','{"accountNumber":"2.YJdlCgdNCC74onPoPh3qCw==|+RgZ4hA+MOoWNYZFMg8o7Q==|MUIMeSMJYMKAzW3YTzth71Z6NlQiQVd5epXIl/9dwcg=","accountType":"2.5GH6z5MsfaZn+ibgWRAqjQ==|tycN6568Wt8mXM0956iKOw==|B/X+nbSxvexxXVcSdRAmdhDV5IiF04wJJtjhUOfFwog=","bankContactPhone":"2.eUDhoL1YaE6pzgM+BxRxdQ==|ZPLdMayspuDqXqwLXMurzQ==|vhy3sMneyCMqVOuV59fCz0sFay6Gm9lTiq43IkvW8Rg=","bankName":"2.4e6m493TCFhc+Z+4CMzfvQ==|GVEVTaE7U2kVg4vK27eP/A==|462oWs0ykD+Tx9EZzmjzbbLAQg7SQw6qqTEp8KrZ/uM=","branchNumber":"2.KgWXmPTf6QCoMrEjfh1/+Q==|pRuxwk6XBD8FbB1r23ktcw==|seT/FL2XBI8X5jMMJQTKivYDeD1wIhauLGD749i+EjU=","iban":"2.V+Iy1lsteObjog6xLiKtXw==|cGdFAGeiP3aAq52nlZ6KOA==|8dlIPo2tqrnHRnC6/8uLSRKjEmaq3k56dgwvys/TXKU=","nameOnAccount":"2.43Bka51mYs6JGbF8oooZqg==|CiNGyi9Q1przjMELMZy7gw==|LuRLfNeqBbo0oDt24Fhc+zH0HpntyCc8ceuhZTgLBW4=","pin":"2.rf6rSy+ExQnzb8r6cTf1vQ==|DUGMsxDFwDrSXlGPi7Wfmg==|fMDpK4WhbgXutHsfZXuECB7JP9785KhMIPIOhhEGq/8=","routingNumber":"2.MinpLN51Z9IYRQHn4Mgb1w==|LNOoe/etHAgiOtYnps23qw==|7gRm4TjpRY9E1jY7uXvmDl3iW/OJ2UlKwlmJp2aEKVQ=","swiftCode":"2.IPKEDqZ2cS9/mtfGABiAug==|r3hNAXnHKa6iKvKm4/jDPQ==|OvItpolBTqDGWYt+QsZbuZI8YTeYRiJugN9e1ff2XRg="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('bcc7f959-90ac-4fd6-b985-4dcba2345b5e','2026-10-25 05:59:03.833041164','2026-10-25 05:59:03.833080123','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,7,'2.DabCTacWJnib4P+ssxeDCA==|pJozyuhogTT6ksVJw7+EFYILeeaz/I4GXk2fnTHAQJE=|87Z8eXOmzZtJL7GOu+ztbnRXMollfUJK66ASWkxhVPA=',NULL,'[]','{"dateOfBirth":"2.i5hHMBjCTk5DYVu3F7uvhw==|s4jyMegjUFMT0fgEKUI+dg==|aAyc/zrMdlYrcRA3fycNGoXNjedx4fwVPDJA8K8tlk8=","expirationDate":"2.+Y4m1CTws3rAg1Vlu1NAsA==|vk08/UVWaI1yVlD82OThQQ==|D8+FMzEWEPyH5OjSslfJUBk27X5lRaEnCm3duH+Qutk=","firstName":"2.vm9zOHf5+XKzKUABfTpjhw==|oOkWCPBTwOIsdudnczWkfw==|KaGqrjxZjW8KkBggOJspwkDnddeRUVtenlY3zxSltNY=","issueDate":"2.LQUFa2NX5eyhNntAqD7/xg==|IHJ2UWit65wZG8ozTdZZQg==|1Tiqh6b8zbTwLklT9Gp8h4KwxDobpfR/2V5YoaYknx4=","issuingAuthority":"2.WebdNAD9IM/a6CjzY4/EvQ==|81qVA0yY78xzbkvfss1IBw==|dkPwg3/VKZnGZex3UxZou7n7XsSea8QPAthPA/Bi12E=","issuingCountry":"2.OmoTyBKSHaXgWHamYhA5Hw==|5BNS8uk/Z4uRk4AEmvLw1g==|p/hEAr3Jkwg1EHG9Db1e2aTCBtvbConAr8GGAXMaBQY=","issuingState":"2.JtaX9rPJUX26BvEpQ2vk8g==|YOQulHdMmaZL4hDc5Zw7uw==|rhPrgq7NnXLB6yAYE2TLVrnteMDfQwpJR0DdAyaKBCU=","lastName":"2.ZN3LU9088hldisoxhK2Mtg==|5nOrybTnTk+IO8LjRDzVtw==|S/5twz5zPzwDYKRCQVykl0JnikI5WfYOUfXE7dAphAk=","licenseClass":"2.gcWfm46J4mV270FdbDgp/w==|1ZQg2vsose2o/eGDE3Q0jg==|B/J07Pp/EHrJMDIaGGcsg03d8UJEAYH7DxE31l3cJVs=","licenseNumber":"2.7WmGCX0adaGUQjjoAuA8nQ==|LTgogy8RIAXHh3qrmuxJIw==|+w+phaqRzUTVb9+Z18G7LK2uorBavtiSIndVB3EqjmU=","middleName":"2.AQo7AVuujtheIcQkLwETmQ==|LI8fR+mTSKRe9vKvlUe2Zg==|lwctL36sINw/dMoms0f/LG3EEmX4VhXVKM0AxTB010w="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('999bfa9b-61c3-478a-af7c-cc06a6a79975','2026-10-25 05:59:03.834116872','2026-10-25 05:59:03.834163830','a5cfb4c1-93db-43e4-9076-136c9667b1d3',NULL,8,'2.ofLLqFKmgID5TkEGdCcwZw==|cckiWjHZzXK7VtQiSY25eg==|n/T+46zL/c4KEdvgnCyBJ0O6y48r3E/eDdjlFLTH4RQ=',NULL,'[]','{"birthPlace":"2.QQbILdONR33l72n/FZKdlA==|FyAx1LNv1P8NNYQqgTBMfA==|vAYrkEw1mOmZp3KoIyCRlURJrviEir8kniKaXZJD0wk=","dateOfBirth":"2.+57zfkno9Tyt+W3t4ZGLLw==|I112CMx30lKbmgQHK6J+ag==|v63GNYPXkzUR/1e//pePxQ+zxHZ/1yJuhBaAOGfKkfA=","expirationDate":"2.yVQ6XaslgulPQiXcnwQ6/w==|UD1uE27WB6kZMApfqstm9w==|3uWaPiZjjZqtCEpIzPeNRIBMEogaN06R+83MXGWVNII=","givenName":"2.kHPqi/ANu7RpcIiXo0InRA==|phsCGT444xXX2p1+H4d+vQ==|7/3Fd6c5WHyTMZmVvQkEWwUfnDSpytGntoRc/tb/Z/U=","issueDate":"2.J8WBbnYUFP2CyTu3sOcOBw==|D6edpNpAnjY2Y9ota7sFEg==|mHK3YI33PMV4Kg8ptM/+p74cewb61eVkrc24lApaIEs=","issuingAuthority":"2.BnhhZXAQca0L1+9cVUAWig==|eQl8UVF3rYaWQpPwydfl7A==|A9DZItsCeCwbUPBzXwi3ElztVSM0AQEhl0OHYmx/Vfs=","issuingCountry":"2.7Nxd+3yMmPY93TmFdRUuhg==|rU0/KdYWprpOaHq2mHgO/w==|s49t22na9xnJftpSUiBL8zNLsC7b+2thgu607wF1rro=","nationalIdentificationNumber":"2.75cvZvvBP7qVRp/k7U89Ug==|ahxNLQ/T92GzpuHsTvXyiA==|KJ0FB4Rm1xJW4PAHIiy91aOAqUPu+lyBxfQr3ONwzyA=","nationality":"2.VMbcojPct1YcFTUvToPNNg==|H2m1qCap4ECLL4WKYYRhfw==|IzmDCGI+azEF5fAA5qUbjesJHrGmJH1kJ4oduOGy/VE=","passportNumber":"2.Gjo1QqcZb5Dqq9BqUtRnpQ==|I73UwvvySdNE2kMAW/LNyA==|LjAuL72tAocxxiVuoCwbPu6lxLCdDFZGrRnfikaJokk=","passportType":"2.HVqAnycmOvtegpELDip2Aw==|f2AqxdCG+riajBc2ablTcw==|jNxnzmEdQ8tZbstuWuHXXWUDahMSGXoD5w4BNIGspd4=","sex":"2.qfvEG5stjKM1fLh9q/ebEw==|0W/F6QrxvmF43Rk3Vj0UOQ==|tFoY6kLUWFBjMd0BvidCqmk9lvoSutKg6Gcakj2xLE0=","surname":"2.xyG/q/m9+KvbaPoFKz73Sw==|XzXYLR8EMCYqqquuDCu9Aw==|udEAD6IVBLLut4OWURzFzSY6v3fV1R/Sm/poOa1e8u8="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('b4cd17a2-9772-4f29-87e3-ea991b02877d','2026-10-25 05:59:04.129842848','2026-10-25 05:59:04.130451265',NULL,'362ce7a0-c564-4031-bab0-a7f418eab5c7',1,'2.ICZz0mxARb5FbEpJAIQSkA==|cIluq47WroA7+pH4X1KMUHxVOg3KySawctzeqaQ0xk0=|UxlxtbV+cg13Vr1BOhSdpR+h4ZCrCBsVQAHeNZQLLbo=','2.qj23AwkDfo7WmLIp6ykskA==|uw2xYM92x+POh/W2vYySwTSARzAJ2bmG9QDsJkX7KZw=|IQe5lySGqEQfdSSEf1NX/Dy0ScKI9vYsE81UeHBco1I=','[]','{"password":"2.QTRkLP2mXCrCVXPG1rairQ==|sA9jf8+InzkIsF+OKbF3Hw==|ogQ/UP6ubO2UNUZIdiuq8xZ5s3HWCKtBSFKLp44wey4=","username":"2.527ubQkCC/OVE6Y4OpHikg==|b7nUUVMfEqUCbaNuqQvLzg==|uiq0dR+RI1BiqNBxkwz4cTiLMQWJc4TwAf/5+0w4+KU="}',NULL,NULL,0,'2.nZZJCu7ek4dMgfuKoutfeA==|eqHjHszJI2jyrSuHsHOvhnpUOUVXZJuwxgPe6Yyj24zlfaMQcJUj6JbEiSNC8C1o4YzEHKWOC0AJqkM6K67Fr8hw/PYYasvAwfqsUsVZ3yU=|Lw7vLMprkFQLr7iW2ma1AFhdFcb3mVEZZ76zE29WqRA=');
INSERT INTO "ciphers" VALUES('863ab098-fdfe-49f4-968e-24c5783c87fd','2026-10-25 05:59:04.131673306','2026-10-25 05:59:04.132037139',NULL,'362ce7a0-c564-4031-bab0-a7f418eab5c7',1,'2.DAzEeYIikMg9Gr4F35mL4Q==|CfmzteMoh9NKxXKfLqg8t3RRMoDMBAV6jWtAtafIWlY=|iYPWsuPTuSbfRTJyJ7z6Eu+Qlv+Utk/P8qyiiqKYmHc=','2.ZxC4CMY7GcjJwdTiwrf2Jg==|T4b8JsSFGk79pKk2JxIyZXoJvCffa9zHrdUJEfspNmw=|WJ5pWZLJN5S8QAe4c9YuP50QFUHSPiJuy8QcI43/ARo=','[]','{"password":"2.KGjSIyzjektoT7mcjk1rVw==|/PUNaaZqeyichiibSBa3gx0h+McTpLf58Z/oCCBsUqg=|JY087k03/FkYEATHkDXvRkhaMmMRlKLrUgbx32tz/Jc=","username":"2.ossP2KyoqTdtAABLSsWAaw==|/k31nWp6iVvXjci6sPruMg==|jvdxSs/OBXGmX57eFapfM2GC0o1cAVsdtgf9I2bx0ZY="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('935fc054-672e-4289-a65d-05a505d125a1','2026-10-25 05:59:04.133036889','2026-10-25 05:59:04.133312389',NULL,'362ce7a0-c564-4031-bab0-a7f418eab5c7',1,'2.KJXlC0E6WNcqYKhoMWBT9Q==|HwS/7eoK64IHs0mKKxsOoRb/534vngArSOPYD7+gHxU=|+RFpuXOoVmbPzRc6lXDLdKujdJdQ3CH1khMsZ1T5qnY=','2.bn81Ej/vq/ceYCnBCTu7HA==|LuT7l0BMzFjzCCPIQtYGyZBT95NEGSO4XSqm/KCHwME=|yXqzetLiMX4ZqyJu614/tZxGYlHnGxkTZ2BCzSTlbWM=','[]','{"password":"2.0O6jXaLi71FKAyMtLfnq6w==|DxrscRnBi3AjuJwujUFgXsUiEMQJDfi5Pas0UOdpCOU=|3E6myi+oIncsPCqkDBtmo+fxswneWvxXNVP1q3Z9WnI=","username":"2.11iTkGHrqgEQSzWn3V++LA==|9zQ/TC6ELQUb8n2rWOWhBg==|UAo2m8XQmw4HlW9+AvVYCMC6/wk6BjzDu2+C2QhMr/k="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('3f8c4146-2276-418d-996b-f513c7373eee','2026-10-25 05:59:04.141613511','2026-10-25 05:59:04.142015552',NULL,'3085313b-dede-4b6e-a8df-e9c9b4c829db',1,'2.ze1KRVKwSrvnxDXWfUIF5g==|y3AYSsATvtOJHrTnAATJb2OWrHP8HYIjiRa2Zs/uiT8=|PkTmfap8WtwjneKaeudjqcLXsNVqZJtJBCfSN1C9OMI=','2.gVQJCUtJnCv5fFSJsKNN6w==|pNNeRcc6ieo6qPVO8LgARC4vflood55Yhre7+Gm9qjU=|VdpDD5vVX2v9s1OlLa4ILhcfPMhe82ZOiDCmLXvRbu0=','[]','{"password":"2.79kd5pa75Uyn1mfJv9hRUw==|8miNqfIBDw5Axq7mmQd5PA==|x5CwONpkIP6/sb99qPlTGMa4lvTCcFNdtCdmJnWGYJ8=","username":"2.NWCRWLCmgFsrHjPAhUaWvg==|+8tw7DEvBy1dhMKo/Er1rQ==|voP53V6rmqTF6jZKDWwYWvCiGxrDxZrWmERN7sLBxLA="}',NULL,NULL,0,'2.RlpFsegENePLSn9swNT+ag==|UbxYxvAOx9ge3g9FNjlNy4qxUUFlgwf+ZcXV92QOWatGunsxyR9AI/eth3/erKoSAc2Z5hUL1RkdV9miydIca2tmw8x3UueL4Br+ybmDMgg=|anAutleZdQLjamDk4OsUSz/etnoJAP1giOmeiHL5i+8=');
INSERT INTO "ciphers" VALUES('80776dc2-efa1-4c0f-9cc8-899a1d3f598b','2026-10-25 05:59:04.143001760','2026-10-25 05:59:04.143296593',NULL,'3085313b-dede-4b6e-a8df-e9c9b4c829db',1,'2.h3sAqGRzsIb9ZctaYJ19Ew==|yEbfT4RRajZV3BPjVtP7inGtxLK/5NGbeb0M5iJh1CU=|li5xY4v1bOz7RuatGp1tCvTee+HlULmebQrp496OTYk=','2.A1FMOEIRK7t00QAaaVwElw==|Av6jJw1mctwrs32ACnlkzwlr5U9MEx+AqbhVxUSlZqg=|1sk/+FvVp7xoQ6DEqXf8F8e5WppjsA3cNcTZi6RHnJQ=','[]','{"password":"2.nN0qBPO6PyBRsKvIU/tvuA==|x815Zr1VKXyZiRDRcxdjVxdHRVy2SkQiJbpQvt5ICsU=|3DqiNgNxNTsoOs9kpddoUDb64C+OKiEVvu2IYV8Ks28=","username":"2.pVnAYGIgJzw+f8Oc+CPYnA==|WcDIskNtjAE/ON4uDpmwLQ==|isuG6NlhkQEu9NNgOIVIAd6k7XKi3QZ24jOi/seovHo="}',NULL,NULL,0,NULL);
INSERT INTO "ciphers" VALUES('4cc6413f-674c-43b6-bc4d-48e0547ca462','2026-10-25 05:59:04.144289385','2026-10-25 05:59:04.144645801',NULL,'3085313b-dede-4b6e-a8df-e9c9b4c829db',1,'2.r20XL2c02FGMcXr2k81fig==|13fzuOkA+Jd0TqX1WkbVQRauTYoBlSkM74aYS1iWGzw=|TR2ACufYgRYCB+UJaIzoKK4BWgfXKS9ZnfTVLmCfOBU=','2.nR5sW7JKIp7fdEURj75xvg==|14f1sx7hqtVOOT+NS/y+SaygoidA7WR6kYKNdmRES/8=|DyxVM7ejoFkXIszQ88tYW9ZDMz7JQb4qQOFe/y//IZM=','[]','{"password":"2.YpOlH8RsB7Srj0wQpxmTTQ==|WqI3eIuXh6sRvVcVfG6ubguZ6iCiYNZkeYFKvLKkJ6k=|hvu8pknf2u9w1C5XQWRnNovKtLK3xo+bjNIsUhU748o=","username":"2.wZ6/Y1QEgx4nNwzV2oKdEg==|vTLBURrfkBUF1Nql6OXbmA==|6jL8u4Y1i9H2Xh0oDYGs0n4CjY1+liqEI5qwCUAkbhY="}',NULL,NULL,0,NULL);
CREATE TABLE ciphers_collections (
  cipher_uuid       TEXT NOT NULL REFERENCES ciphers (uuid),
  collection_uuid TEXT NOT NULL REFERENCES collections (uuid),
  PRIMARY KEY (cipher_uuid, collection_uuid)
);
INSERT INTO "ciphers_collections" VALUES('b4cd17a2-9772-4f29-87e3-ea991b02877d','987aa335-7e33-4925-8b4d-600ebcfaf081');
INSERT INTO "ciphers_collections" VALUES('b4cd17a2-9772-4f29-87e3-ea991b02877d','06ce75b7-78b8-4d51-8ca7-22df5b1377cb');
INSERT INTO "ciphers_collections" VALUES('863ab098-fdfe-49f4-968e-24c5783c87fd','987aa335-7e33-4925-8b4d-600ebcfaf081');
INSERT INTO "ciphers_collections" VALUES('935fc054-672e-4289-a65d-05a505d125a1','06ce75b7-78b8-4d51-8ca7-22df5b1377cb');
INSERT INTO "ciphers_collections" VALUES('3f8c4146-2276-418d-996b-f513c7373eee','69662946-b194-4ea4-bf78-53bd7c316246');
INSERT INTO "ciphers_collections" VALUES('3f8c4146-2276-418d-996b-f513c7373eee','3b668be0-c582-425c-9f51-f7e8a700df30');
INSERT INTO "ciphers_collections" VALUES('80776dc2-efa1-4c0f-9cc8-899a1d3f598b','69662946-b194-4ea4-bf78-53bd7c316246');
INSERT INTO "ciphers_collections" VALUES('4cc6413f-674c-43b6-bc4d-48e0547ca462','3b668be0-c582-425c-9f51-f7e8a700df30');
CREATE TABLE collections (
  uuid     TEXT NOT NULL PRIMARY KEY,
  org_uuid TEXT NOT NULL REFERENCES organizations (uuid),
  name     TEXT NOT NULL
, external_id TEXT);
INSERT INTO "collections" VALUES('987aa335-7e33-4925-8b4d-600ebcfaf081','362ce7a0-c564-4031-bab0-a7f418eab5c7','2.Dz7SJ6Nd6FTYr+jWRyXmcA==|a5hAEYe87aXDwJqLXlTSmg==|Hdw09RN05J/Nc2FfGreUJW1w+osowfp2BzEngQJrAXk=',NULL);
INSERT INTO "collections" VALUES('06ce75b7-78b8-4d51-8ca7-22df5b1377cb','362ce7a0-c564-4031-bab0-a7f418eab5c7','2.JIiOkcTWjp+rqHp/QTi1Zw==|HGv4VpYxMb7hVkTTy7FQTQ==|It6i5D0kDdFx0flotzLu7W5NuuJ57zz1aMeuoQ6Ys9k=',NULL);
INSERT INTO "collections" VALUES('69662946-b194-4ea4-bf78-53bd7c316246','3085313b-dede-4b6e-a8df-e9c9b4c829db','2.jsRPvWNZq6tjnCDeBNkJVA==|ssZk5YudeGVdc3hlbO0G2A==|QFmnUAIuZBijVi1a5/aZOdFiZ+8I0m0lg6xcNy4ui0U=',NULL);
INSERT INTO "collections" VALUES('3b668be0-c582-425c-9f51-f7e8a700df30','3085313b-dede-4b6e-a8df-e9c9b4c829db','2.eRgBvyPheDaq65wctVYlzQ==|IklvICW7CtXcJjODCbBETQ==|fXhh0Eyi8RwYx7R61Nt5oO3i7Q+WEZiCaZ13fiEChdg=',NULL);
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
INSERT INTO "devices" VALUES('30000000-0000-4000-8000-000000000001','2026-10-25 05:59:03.810848089','2026-10-25 05:59:04.440653735','a5cfb4c1-93db-43e4-9076-136c9667b1d3','bwb-integration',8,NULL,'2pcfophyuormGdKWrW05yduf9yYZjqcC1E3u9Mk29BaPcW79F8FHyOpvNj3uZxLkWw_UKIQoCYFYayk-3C_ZOQ==',NULL,'52c74c65-c3b1-4f02-9303-f2228a32475e');
INSERT INTO "devices" VALUES('30000000-0000-4000-8000-000000000001','2026-10-25 05:59:04.611234715','2026-10-25 05:59:04.611399215','e2d3677c-568a-42fd-ba18-7c5315d37c49','bwb-integration',8,NULL,'f6hitqyLLAPHaOR3uI7-7cTaV0ZazhFTowO8aUXW68OIWVnF4wpV2XKjSZYax3l4zE0JTr-3bIU8EQ6xyydxFA==',NULL,'3689721f-2b2c-4896-9aff-8e0719785944');
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
INSERT INTO "favorites" VALUES('a5cfb4c1-93db-43e4-9076-136c9667b1d3','306875da-d69b-4625-b177-fea2df37f351');
CREATE TABLE folders (
  uuid       TEXT     NOT NULL PRIMARY KEY,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  user_uuid  TEXT     NOT NULL REFERENCES users (uuid),
  name       TEXT     NOT NULL
);
INSERT INTO "folders" VALUES('57e80c20-8fa1-4147-8844-93b13f41572b','2026-10-25 05:59:03.816165670','2026-10-25 05:59:03.816231170','a5cfb4c1-93db-43e4-9076-136c9667b1d3','2.Kf4eO3w1raFCtBffaNAHFA==|kHo0HwozNIlZ9LxwDq+hkw==|+OfskuuTqQ5Oj+daAZOb7sFM3G5QjuQG4PAReyfgJaU=');
INSERT INTO "folders" VALUES('8b01cf56-d604-44bd-acd5-92eb0a9a54a3','2026-10-25 05:59:03.817358837','2026-10-25 05:59:03.817450462','a5cfb4c1-93db-43e4-9076-136c9667b1d3','2.AXKXzoEPlGGJBjLqXUpz3w==|28m82DDMedzlXuc7MdW//w==|7+VnzX4BvRO7RFOw3U2L6ORnd1Y2LgFbM7I3RUd53h8=');
CREATE TABLE folders_ciphers (
  cipher_uuid TEXT NOT NULL REFERENCES ciphers (uuid),
  folder_uuid TEXT NOT NULL REFERENCES folders (uuid),

  PRIMARY KEY (cipher_uuid, folder_uuid)
);
INSERT INTO "folders_ciphers" VALUES('306875da-d69b-4625-b177-fea2df37f351','57e80c20-8fa1-4147-8844-93b13f41572b');
INSERT INTO "folders_ciphers" VALUES('37296b7b-354e-4898-a4d6-57fb085fe9b8','57e80c20-8fa1-4147-8844-93b13f41572b');
INSERT INTO "folders_ciphers" VALUES('aee95654-6917-45ec-b1f6-cd4f853701fd','8b01cf56-d604-44bd-acd5-92eb0a9a54a3');
INSERT INTO "folders_ciphers" VALUES('77d96bf1-9b85-47e8-baf9-c9ea01604c50','57e80c20-8fa1-4147-8844-93b13f41572b');
INSERT INTO "folders_ciphers" VALUES('fa63206b-ef23-4152-a2a2-b95ecc04a047','57e80c20-8fa1-4147-8844-93b13f41572b');
INSERT INTO "folders_ciphers" VALUES('61368e44-4751-4cc0-917f-e78fb7aba727','57e80c20-8fa1-4147-8844-93b13f41572b');
INSERT INTO "folders_ciphers" VALUES('a390138f-dc99-4969-a1f5-ea9b8da03a5b','57e80c20-8fa1-4147-8844-93b13f41572b');
INSERT INTO "folders_ciphers" VALUES('20160329-5a43-4050-8080-74b54d13f21b','57e80c20-8fa1-4147-8844-93b13f41572b');
INSERT INTO "folders_ciphers" VALUES('b4cd17a2-9772-4f29-87e3-ea991b02877d','57e80c20-8fa1-4147-8844-93b13f41572b');
INSERT INTO "folders_ciphers" VALUES('3f8c4146-2276-418d-996b-f513c7373eee','57e80c20-8fa1-4147-8844-93b13f41572b');
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
INSERT INTO "organizations" VALUES('362ce7a0-c564-4031-bab0-a7f418eab5c7','DEMO Engineering','bwb-integration@example.invalid',NULL,NULL);
INSERT INTO "organizations" VALUES('3085313b-dede-4b6e-a8df-e9c9b4c829db','DEMO Research','bwb-integration@example.invalid',NULL,NULL);
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
INSERT INTO "users" VALUES('a5cfb4c1-93db-43e4-9076-136c9667b1d3','2026-10-25 05:59:03.578866506','2026-10-25 05:59:04.144762801','bwb-integration@example.invalid','BWB integration',X'DB11AFB9641A518F2976DF4F7E0D6BA7E00605A971D57AD5E5141379B06E9BA8',X'E80863BB0AC07BBF15A431321AC3F9DC983E361B090E25FE04D54D4F432A8D4C8DDFE3770980495E93173762DF6A899B6758AE44A37E2FCE9C7F49C2BB88BBB6',600000,NULL,'2.vdzKGpTrWcNrudwWNk0GZw==|aAzsS5MSq509FOY+exN+knypBtWPqX8lPxhW+Us52FHhc5eIj0EWoyaTo51WQWO/7o146ZxlaiQFS77kQ+aF0hiweXWvzkWt7OTOOjZ/HyU=|5T3FwJuY7/S/+J7UidNngoYTvwW5CtA0CuK+rSqcWtw=','2.b/AkHOErr55aqr6oVTS5XA==|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',NULL,NULL,'e5c34276-999a-4ad6-9aae-f1ff6428a41c','[]','[]',0,600000,NULL,NULL,0,NULL,NULL,1,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO "users" VALUES('e2d3677c-568a-42fd-ba18-7c5315d37c49','2026-10-25 05:59:04.039007756','2026-10-25 05:59:04.144742801','bwb-restricted@example.invalid','bwb-restricted@example.invalid',X'EA3F78EF6F403E5C2304FFFA7F099E274E5AB41C1D8AD98C5FB50BA8C4880F47',X'C68FCBCFDD590CE55F1EB4701BCA1909B8DF4D41B5A27F69B5340DF9DB265F850509EC4B9B9540CA0F77C4872A671C0D83CC288FCD562DBDB5E704884310F98B',600000,NULL,'2.GS+jnZ7U6n5Fi3+WruWX3w==|bPCUJUJIEK6go0XBtUfij0QCziOP5OePulhAmoPnxjSLflIuAqdrWefJwuh6OGbfrDclz4tnr1ZoqA0m5u5M7UCuzhHx46LiOkI9bFf+UJI=|Mi44a+glPPXHGGEiKnDq2gtA+KltKYjNpCIEytW3iQM=','2.uSFlIpUUYsJph50bIhpw7Q==|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',NULL,NULL,'aab03e5e-9e0b-447f-8dad-4efae5aa2c96','[]','[]',0,600000,NULL,NULL,0,NULL,NULL,1,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
CREATE TABLE users_collections (
  user_uuid       TEXT NOT NULL REFERENCES users (uuid),
  collection_uuid TEXT NOT NULL REFERENCES collections (uuid), read_only BOOLEAN NOT NULL DEFAULT 0, hide_passwords BOOLEAN NOT NULL DEFAULT 0, manage BOOLEAN NOT NULL DEFAULT 0,
  PRIMARY KEY (user_uuid, collection_uuid)
);
INSERT INTO "users_collections" VALUES('e2d3677c-568a-42fd-ba18-7c5315d37c49','987aa335-7e33-4925-8b4d-600ebcfaf081',1,1,0);
INSERT INTO "users_collections" VALUES('e2d3677c-568a-42fd-ba18-7c5315d37c49','06ce75b7-78b8-4d51-8ca7-22df5b1377cb',1,0,0);
INSERT INTO "users_collections" VALUES('e2d3677c-568a-42fd-ba18-7c5315d37c49','69662946-b194-4ea4-bf78-53bd7c316246',1,1,0);
INSERT INTO "users_collections" VALUES('e2d3677c-568a-42fd-ba18-7c5315d37c49','3b668be0-c582-425c-9f51-f7e8a700df30',1,0,0);
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
INSERT INTO "users_organizations" VALUES('977138ff-e9d6-4736-ba6a-bf3a02c58ecc','a5cfb4c1-93db-43e4-9076-136c9667b1d3','362ce7a0-c564-4031-bab0-a7f418eab5c7',1,'4.0KJr/Hg1efXQunDHssEqBUIi5bHgoKwDJVPvZCA6nq5hKEhYL8e/34NRItJUV997KG9NusVr4eciUXBKYCIHFu1YSQPJiD244sCk/C5dWH2MkaRNbWetbVjp1w3rOdCuKbm8pN8Abvp0/lPOYC4K/+XUiozjmiZe7QfGmgPfLLoB7R39/ypLjXitGrHrkyZ/SrmehMLBRG+U+tQEabCiNWE1134tYH6rIVLeHfZdxwvGKEX7tZjcObpZVcY2mPzjxXuvB4x1x2nMb54ENPldm6ObUq7HA/5sQo+fQJi58bL1LNxSasoqwmdAQigedlTlZ+6Jgy9AdjJ7itxZb3VFyQ==',2,0,NULL,NULL,NULL);
INSERT INTO "users_organizations" VALUES('a7680f42-5769-456a-ae6e-310b6e40a76f','e2d3677c-568a-42fd-ba18-7c5315d37c49','362ce7a0-c564-4031-bab0-a7f418eab5c7',0,'4.VL8wMI3tuE3QV4DYYbHg3cjtqwAsyJ4OPs82uQQw3ac09R2AlarYlm6oaaootPrTwwA+nh3wfSEjONrmlLis1OvbJh9Za75l4GihE5LBzkdY3g9FPbHbjxJprkuX43VKEltEXr8zpsBIushhxIZLaLFGbluI1yp02Vf+T1A+jvHbWZxtD5WE2O/VJEu0BtSpVefkHsGkxVIXDFC44n35na9z2bkv/ebJkLzJ4pGo5dr6wGAdgCpMWgzd82LQ5xoBPDNTf21u1v2gDY9j8tXG4G1+WBHrI/LF7mWC0hHyU3iPr/dCV55CreDRakvo3VHrtJm+3jG2VQPF5xY6ziQh8Q==',2,2,NULL,NULL,'bwb-integration@example.invalid');
INSERT INTO "users_organizations" VALUES('b0a5951c-81c0-4d92-858a-b1e6eb37fa7f','a5cfb4c1-93db-43e4-9076-136c9667b1d3','3085313b-dede-4b6e-a8df-e9c9b4c829db',1,'4.SqswVKg2xu5EwiLvn7EoQgiRK8pmsV8QiLuPz6vrrqTh8wi8YfMOB1+u2cVI1DLF/1WV8kbsv3GBaxQ8CW0Yvbd7F8XNL/svwFzxogcaiQx+niWUs3elIT2FqIXndgB5mM0eFoCcesoEPJcvX2UXYuUZG74PYwirlksqGm3d3mzbzKiy02lNvz9oGcODkzhVStHhpSb4zzbzqT3wbuHg1jFtpb/vqKTQVKhTeVkpABSm0o+yDOZMH+1s+yq3+2WITYpz9w6K1S8aN4tYKvLR7A76WF/Y3q2ZF6AtvbfZNDe0YCA4ke25Ztun2cyhbJcptykt0lnIaebezGsnRXwfqw==',2,0,NULL,NULL,NULL);
INSERT INTO "users_organizations" VALUES('09bb002a-d92f-4192-9478-e48b0a998a26','e2d3677c-568a-42fd-ba18-7c5315d37c49','3085313b-dede-4b6e-a8df-e9c9b4c829db',0,'4.vanD8DvcJ/PJfaZNbuS1Vi2JKkxacIAbnkmQAC3Pq+mi+6HICtoQVMnU7N+Kg7Ia2XyI01g+Lw4EwgxDPGosghu20byqqxiR4eKj2DOzMQVbQfo9YB6AdaQpABddkO23DEG0zA6+vQ7S088bjq1BnVqFNbqqP/yystJSCwM49kF+4SvmRr+E2H5NQU+ghmXdyuTmmzCT990CHG2ZxodiM1KbUbEaWi/vn8CPyU3CURDPRo7E71dXSjjcgh1EiXRlruC/o4n2NNLlaYPnuKYAuKEtnzldldGpJULBH3NPQn1d/MsaMJ2tFHWtqte2mBSk3r5PgDuajNgHphB4oY/9fg==',2,2,NULL,NULL,'bwb-integration@example.invalid');
COMMIT;
