#!python3
import argparse
import dataclasses
import json
import logging
from pathlib import Path
import sys
from typing import Any, List, Mapping, Optional

@dataclasses.dataclass
class Column:
	name: str
	type: str

def main() -> None:
	parser = argparse.ArgumentParser()
	parser.add_argument("src", type=Path, help="The directory to pull schema from")
	parser.add_argument("dest", type=Path, help="The output file to write to")
	args = parser.parse_args()

	logging.basicConfig(level=logging.DEBUG)

	# Parse all the schemas
	schema: Mapping[str, List[Column]] = {}
	for src in args.src.glob("*.json"):
		table_name = src.stem
		print(table_name)
		with open(src, "r") as f:
			data = json.loads(f.read())
		if "fields" not in data:
			continue
		columns = read_schema(data)
		schema[table_name] = columns
	
	with open(args.dest, "w") as output:
		output.write("-- +goose Up\n")
		# Write out what we parsed
		for table_name in sorted(schema):
			columns = schema[table_name]
			output.write(f"CREATE TABLE FS_{table_name} (\n\t")
			lines = ",\n\t".join(f"{c.name} {c.type}" for c in sorted(columns, key=lambda c: c.name))
			output.write(lines)
			output.write(");\n\n")
		output.write("-- +goose Down\n")
		for table_name in sorted(schema):
			output.write(f"DROP TABLE {table_name};\n")

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
			type_ = "INTEGER PRIMARY KEY"
		else:
			type_ = column_type(field["type"])
		columns.append(Column(
			name=field["name"],
			type=type_,
		))
	return columns
		
		
def column_type(name: str, additional: Optional[str] = None) -> str:
	result = ""
	if name == "esriFieldTypeInteger":
		result = "INT8"
	elif name == "esriFieldTypeDate":
		result = "BIGINT"
	elif name == "esriFieldTypeDouble":
		result = "DOUBLE PRECISION"
	elif name == "esriFieldTypeInteger":
		result = "INT8"
	elif name == "esriFieldTypeSmallInteger":
		result = "INT2"
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
