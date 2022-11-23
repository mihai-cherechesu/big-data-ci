from schema import Schema

def main():
    print(Schema(int).validate(123))

if __name__ == "__main__":
    main()