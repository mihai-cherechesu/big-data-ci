FROM python:3.9-alpine

WORKDIR /app

COPY parser.py ./
COPY requirements.txt ./

RUN pip3 install \
    -r requirements.txt

EXPOSE 8082

CMD ["python3", "/app/parser.py"]