plugins {
    id("java")
    alias(libs.plugins.intelliJPlatform)
}

group = providers.gradleProperty("pluginGroup").get()
version = providers.gradleProperty("pluginVersion").get()

repositories {
    mavenCentral()
    intellijPlatform {
        defaultRepositories()
    }
}

dependencies {
    implementation(libs.snakeyaml)

    testImplementation(libs.junit)
    testImplementation(libs.mockito.core)

    intellijPlatform {
        intellijIdea(providers.gradleProperty("platformVersion"))
        bundledPlugin("org.jetbrains.plugins.yaml")
        testFramework(org.jetbrains.intellij.platform.gradle.TestFrameworkType.Platform)
    }
}

intellijPlatform {
    pluginConfiguration {
        version = providers.gradleProperty("pluginVersion")
        description = """
            Generate sample YAML files from Kubernetes Custom Resource Definitions.

            Right-click on CRD YAML files to access the CRD to Sample YAML menu.
            Generate complete, minimal, or commented samples. Validate existing samples against CRD schemas.
        """.trimIndent()

        ideaVersion {
            sinceBuild = providers.gradleProperty("pluginSinceBuild")
            untilBuild = providers.gradleProperty("pluginUntilBuild")
        }
    }

    pluginVerification {
        ides {
            recommended()
        }
    }

    signing {
        certificateChain = providers.environmentVariable("CERTIFICATE_CHAIN")
        privateKey = providers.environmentVariable("PRIVATE_KEY")
        password = providers.environmentVariable("PRIVATE_KEY_PASSWORD")
    }

    publishing {
        token = providers.environmentVariable("PUBLISH_TOKEN")
    }
}

java {
    // 2026.2 ships Java 25 class files, so the compiler has to be at least that new...
    toolchain {
        languageVersion = JavaLanguageVersion.of(25)
    }
}

tasks.withType<JavaCompile>().configureEach {
    // ...but the plugin itself targets 21 so it still runs on the IDEs down at pluginSinceBuild.
    options.release = 21
    options.compilerArgs.add("-Xlint:deprecation")
}
