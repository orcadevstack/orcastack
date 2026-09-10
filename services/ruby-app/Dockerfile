
FROM ruby:3.2-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y \
    build-essential \
    libpq-dev \
    git \
    && rm -rf /var/lib/apt/lists/*

COPY Gemfile Gemfile.lock ./

RUN gem install bundler && bundle install

COPY . .


FROM ruby:3.2-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app /app

EXPOSE 3000

CMD ["ruby", "server.rb"]
