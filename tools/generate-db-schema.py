#!python3
import argparse
import dataclasses
import json
import logging
import sys
from typing import Any, List, Mapping, Optional

@dataclasses.dataclass
class Column:
	name: str
	type: str

def main() -> None:
	parser = argparse.ArgumentParser()
	parser.add_argument("tablename", help="The name of the table")
	parser.add_argument("input", help="The JSON file to ingest")
	args = parser.parse_args()

	logging.basicConfig(level=logging.DEBUG)
	with open(args.input, "r") as f:
		data = json.loads(f.read())

	columns = read_schema(data)
	sys.stdout.write(f"CREATE TABLE FS_{args.tablename} (\n\t")
	lines = ",\n\t".join(f"{c.name} {c.type}" for c in sorted(columns, key=lambda c: c.name))
	print(lines)
	print(")")

def read_schema(data: Mapping[str, Any]) -> List[Column]:
	columns = [
		Column(
			name="geometry_x",
			type="FLOAT",
		),
		Column(
			name="geometry_y",
			type="FLOAT",
		),
	]
			
	for field in data["fields"]:
		if field["name"] == data["uniqueIdField"]["name"]:
			type_ = column_type(field["type"], "PRIMARY KEY")
		else:
			type_ = column_type(field["type"])
		columns.append(Column(
			name=field["name"],
			type=type_,
		))
	return columns
		
		
def column_type(name: str, additional: Optional[str] = None) -> str:
	result = ""
	if name == "esriFieldTypeDate":
		result = "BIGINT"
	elif name == "esriFieldTypeInteger":
		result = "INTEGER"
	elif name == "esriFieldTypeSmallInteger":
		result = "INTEGER"
	elif name == "esriFieldTypeString":
		result = "TEXT"
	elif name == "esriFieldTypeGlobalID":
		result = "TEXT"
	elif name == "esriFieldTypeGUID":
		result = "TEXT"
	elif name == "esriFieldTypeOID":
		result = "TEXT"
	else:
		raise Exception(f"Not sure how to translate {name}")
	return result + " " + additional if additional else result

if __name__ == "__main__":
	main()
