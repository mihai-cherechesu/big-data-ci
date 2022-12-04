from schema import Schema, And, SchemaError, Optional
from flask import Flask, request, Response, json
from werkzeug.exceptions import BadRequest, NotFound


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


app = Flask(__name__)
parser = Parser()


def main():
    print("starting to listen on :8082")
    app.run(host="0.0.0.0", port=8082)


@app.route("/parse", methods=["POST"])
def parse_pipeline():
    try:
        parser.schema.validate(request.get_json())
    except SchemaError as e:
        return Response("parsing failed, reason: " + str(e), status=500)

    return Response("parsing successful", status=200)


@app.route("/health", methods=["GET"])
def check_health():
    return Response("parser is healthy", status=200)


if __name__ == "__main__":
    main()
