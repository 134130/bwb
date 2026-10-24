"""Check the synthetic import fixture without reading any real vault."""
import json
from pathlib import Path
import unittest
import uuid


class FixtureTest(unittest.TestCase):
    def test_import_shape_and_scenarios(self):
        data = json.loads((Path(__file__).resolve().parents[1] / "testdata/bitwarden.json").read_text())
        self.assertIs(data["encrypted"], False)
        folders, items = data["folders"], data["items"]
        self.assertEqual(len(folders), 2)
        self.assertEqual(len(items), 13)
        ids = [x["id"] for x in folders + items]
        self.assertEqual(len(ids), len(set(ids)))
        for value in ids:
            self.assertEqual(str(uuid.UUID(value)), value)
        folder_ids = {x["id"] for x in folders}
        self.assertEqual({x["type"] for x in items}, set(range(1, 9)))
        for item in items:
            self.assertIn(item["folderId"], folder_ids | {None})
            self.assertIsNone(item["organizationId"])
            self.assertEqual(item["collectionIds"], [])
            self.assertIn({1: "login", 2: "secureNote", 3: "card", 4: "identity", 5: "sshKey", 6: "bankAccount", 7: "driversLicense", 8: "passport"}[item["type"]], item)
        self.assertEqual(sum(x["name"] == "Database" for x in items), 2)
        self.assertTrue(any(x.get("login", {}).get("password") == "" for x in items))
        self.assertTrue(any("\n" in (x["notes"] or "") for x in items))
        self.assertTrue(any(not x["name"].isascii() for x in items))
        self.assertTrue(any(f["type"] == 1 for x in items for f in x["fields"]))
        self.assertTrue(any(x.get("login", {}).get("totp") for x in items))


if __name__ == "__main__":
    unittest.main()
