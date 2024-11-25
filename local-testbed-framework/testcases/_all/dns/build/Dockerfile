FROM python:3.11-slim

WORKDIR /app

COPY build/requirements.txt .
RUN pip install -r requirements.txt

COPY src/ .

COPY build/docker-entrypoint.sh .
RUN chmod +x ./docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
