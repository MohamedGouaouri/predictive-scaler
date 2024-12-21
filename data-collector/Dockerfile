FROM python:3.8

WORKDIR /app

COPY . /app

RUN pip install --no-cache-dir \
    networkx==3.0 \
    numpy==1.24.0 \
    pika==1.3.0 \
    torch\
    torch-geometric==2.4.0

CMD ["python", "app.py"]
