from schema import Schema, And, SchemaError, Optional


class Parser:
    def __init__(self) -> None:
        self.schema = Schema({
            "image": str,
            And(str, lambda s: s not in ("image")): {
                "script": [str],
                Optional("depends_on"): [str],
                Optional("artifacts"): [str]
            }
        })


def main():
    print("initializing parser")
    parser = Parser()


if __name__ == "__main__":
    main()
