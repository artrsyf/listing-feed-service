FROM sbtscala/scala-sbt:eclipse-temurin-alpine-21.0.8_9_1.12.9_2.12.21 AS build

WORKDIR /app

COPY project ./project
COPY build.sbt .
COPY src ./src
COPY secrets.env .

RUN sbt stage

FROM eclipse-temurin:21-jre

WORKDIR /app

# копируем только результат сборки
COPY --from=build /app/target/universal/stage/ .

EXPOSE 8080

ENTRYPOINT ["bin/feed"]