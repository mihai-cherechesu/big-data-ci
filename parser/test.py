from schema import Schema, And, SchemaError, Optional
from parser import Parser

parser = Parser()


tests = [
    {
        "build": "build",
    },
    {
        "image": "image",
        "build": "build",
    },
    {
        "image": "image",
        "build": {}
    },
    {
        "image": "image",
        "build": {
            "script": "wrong script"
        }
    },
    {
        "image": "image",
        "build": {
            "script": ["good", "script"]
        }
    },
    {
        "image": "image",
        "build": {
            "script": ["good", "script"],
        },
        "test": {
            "script": ["test", "binary"],
            "depends": ["bad", "depends"]
        }
    },
    {
        "image": "image",
        "build": {
            "script": ["good", "script"],
        },
        "test": {
            "script": ["test", "binary"],
            "depends_on": ["build"]
        }
    }
]

for test in tests:
    try:
        parser.schema.validate(test)
        print("Passed!")
    except SchemaError as e:
        print("Failed: " + str(e))
