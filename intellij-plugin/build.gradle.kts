plugins {
    id("java")
    alias(libs.plugins.kotlin)
    alias(libs.plugins.intelliJPlatform)
}

group = "com.skarlso"
version = "1.0.2"

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
    testImplementation(libs.mockito.inline)
    
    intellijPlatform {
        intellijIdeaCommunity("2025.1.3")
        bundledPlugins("com.intellij.java", "org.jetbrains.plugins.yaml")
        testFramework(org.jetbrains.intellij.platform.gradle.TestFrameworkType.Platform)
    }
}

intellijPlatform {
    pluginConfiguration {
        version = providers.gradleProperty("pluginVersion").orElse("1.0.2")
        description = """
            Generate sample YAML files from Kubernetes Custom Resource Definitions.
            
            Right-click on CRD YAML files to access the CRD to Sample YAML menu. 
            Generate complete, minimal, or commented samples. Validate existing samples against CRD schemas.
        """.trimIndent()
        
        ideaVersion {
            sinceBuild = "251"
            untilBuild = "251.*"
        }
    }
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
}
