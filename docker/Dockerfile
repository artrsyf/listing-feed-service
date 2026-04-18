FROM sbtscala/scala-sbt:eclipse-temurin-21.0.2_13_1.10.2 AS build

WORKDIR /app

COPY project ./project
COPY build.sbt .

COPY src ./src

RUN sbt clean assembly

FROM eclipse-temurin:21-jre

WORKDIR /app

# копируем только результат сборки
COPY --from=build /app/target/scala-2.13/*assembly*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]