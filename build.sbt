ThisBuild / version := "0.1.0-SNAPSHOT"

ThisBuild / scalaVersion := "2.13.18"

ThisBuild / envFileName := "secrets.env"

val tapirVersion = "1.13.15"

lazy val root = (project in file("."))
  .enablePlugins(JavaAppPackaging)
  .settings(
    name := "feed",
    libraryDependencies ++= Seq(
      "dev.zio" %% "zio" % "2.1.25",
      "dev.zio" %% "zio-logging" % "2.5.3",
      "dev.zio" %% "zio-logging-slf4j" % "2.5.3",
      "org.slf4j" % "slf4j-api" % "2.0.17",
      "ch.qos.logback" % "logback-classic" % "1.5.32",
      "dev.zio" %% "zio-http" % "3.10.1",
      "dev.zio" %% "zio-json" % "0.7.44",
      "com.softwaremill.sttp.tapir" %% "tapir-core" % tapirVersion,
      "com.softwaremill.sttp.tapir" %% "tapir-zio-http-server" % tapirVersion,
      "com.softwaremill.sttp.tapir" %% "tapir-json-zio" % tapirVersion,
      "com.softwaremill.sttp.tapir" %% "tapir-openapi-docs" % tapirVersion,
      "com.softwaremill.sttp.tapir" %% "tapir-swagger-ui-bundle" % tapirVersion,
      "org.tpolecat" %% "doobie-core" % "1.0.0-RC12",
      "org.tpolecat" %% "doobie-postgres" % "1.0.0-RC12",
      "org.tpolecat" %% "doobie-hikari" % "1.0.0-RC12",
      "dev.zio" %% "zio-interop-cats" % "23.1.0.13",
      "org.postgresql" % "postgresql" % "42.7.10",
      "dev.zio" %% "zio-config" % "4.0.7",
      "dev.zio" %% "zio-config-typesafe" % "4.0.7",
      "dev.zio" %% "zio-config-magnolia" % "4.0.7",
      "io.scalaland" %% "chimney" % "1.9.0",
      "com.sksamuel.elastic4s" %% "elastic4s-client-esjava" % "8.11.5",
      "com.sksamuel.elastic4s" %% "elastic4s-effect-zio" % "8.11.5",
      "com.beachape" %% "enumeratum-doobie" % "1.9.7",
      "io.minio" % "minio" % "8.5.7"
    ),
    dependencyOverrides ++= Seq(
      "dev.zio" %% "zio-json" % "0.7.44",
      "com.fasterxml.jackson.core" % "jackson-databind" % "2.14.3",
      "com.fasterxml.jackson.core" % "jackson-core" % "2.14.3",
      "com.fasterxml.jackson.core" % "jackson-annotations" % "2.14.3"
    )
  )
