FROM public.ecr.aws/docker/library/python:3.9-slim

LABEL maintainer=opomer@amazon.co.uk

WORKDIR /app
COPY app/requirements.txt .

RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

COPY app/main.py .

ENTRYPOINT ["python"]
CMD ["main.py"]