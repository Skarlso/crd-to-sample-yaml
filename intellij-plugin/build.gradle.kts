plugins {
    id("java")
    id("org.jetbrains.kotlin.jvm") version "2.0.21"
    id("org.jetbrains.intellij") version "1.17.3"
}

group = "com.skarlso"
version = "1.0.0"

repositories {
    mavenCentral()
}

intellij {
    version.set("2025.1.3")
    type.set("IC") // IntelliJ IDEA Community Edition
    
    plugins.set(listOf(
        "com.intellij.java",
        "org.jetbrains.plugins.yaml"
    ))
}

tasks {
    withType<JavaCompile> {
        sourceCompatibility = "17"
        targetCompatibility = "17"
    }
    
    withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
        compilerOptions {
            jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
        }
    }

    patchPluginXml {
        sinceBuild.set("251")
        untilBuild.set("251.*")
    }

    signPlugin {
        certificateChain.set(System.getenv("CERTIFICATE_CHAIN"))
        privateKey.set(System.getenv("PRIVATE_KEY"))
        password.set(System.getenv("PRIVATE_KEY_PASSWORD"))
    }

    publishPlugin {
        token.set(System.getenv("PUBLISH_TOKEN"))
    }
    
    // Disable problematic buildSearchableOptions task for newer versions
    buildSearchableOptions {
        enabled = false
    }
}

dependencies {
    implementation("org.yaml:snakeyaml:2.2")
    
    testImplementation("junit:junit:4.13.2")
    testImplementation("org.mockito:mockito-core:5.7.0")
    testImplementation("org.mockito:mockito-inline:5.2.0")
}
