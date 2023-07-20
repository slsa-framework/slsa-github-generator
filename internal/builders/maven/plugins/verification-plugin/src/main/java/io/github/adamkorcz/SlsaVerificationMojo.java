// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package io.github.adamkorcz;

import org.apache.maven.artifact.Artifact;
import org.apache.maven.execution.MavenSession;
import org.apache.maven.plugin.AbstractMojo;
import org.apache.maven.plugin.BuildPluginManager;
import org.apache.maven.plugin.MojoExecutionException;
import org.apache.maven.plugin.MojoFailureException;
import org.apache.maven.plugins.annotations.LifecyclePhase;
import org.apache.maven.plugins.annotations.Component;
import org.apache.maven.plugins.annotations.Mojo;
import org.apache.maven.plugins.annotations.Parameter;
import org.apache.maven.project.MavenProject;

import static org.twdata.maven.mojoexecutor.MojoExecutor.*;

import java.io.File;
import java.util.Set;

/*
 SlsaVerificationMojo is a Maven plugin that wraps https://github.com/slsa-framework/slsa-verifier.
 At a high level, it does the following:
   1: Install the slsa-verifier.
   2: Loop through all dependencies of a pom.xml. Resolve each dependency.
   3: Check if each dependency also has a provenance file.
   4: Run the slsa-verifier for the dependency if there is a provenance file.
   5: Output the results.

 The plugin is meant to be installed and then run from the root of a given project file.
 A pseudo-workflow looks like this:
   1: git clone --depth=1 https://github.com/slsa-framework/slsa-github-generator
   2: cd slsa-github-generator/internal/builders/maven/plugins/verification-plugin
   3: mvn clean install
   4: cd /tmp
   5: git clone your repository
   6: cd into your repository
   7: mvn io.github.adamkorcz:slsa-verification-plugin:0.0.1:verify
*/
@Mojo(name = "verify", defaultPhase = LifecyclePhase.VALIDATE)
public class SlsaVerificationMojo extends AbstractMojo {
    @Parameter(defaultValue = "${project}", required = true, readonly = true)
    private MavenProject project;

    /**
      * Custom path of GOHOME, default value is $HOME/go
    **/
    @Parameter(property = "slsa.verifier.path", required = true)
    private String verifierPath;

    @Component
    private MavenSession mavenSession;

    @Component
    private BuildPluginManager pluginManager;


    public void execute() throws MojoExecutionException, MojoFailureException {
        // Verify the slsa of each dependency
        Set<Artifact> dependencyArtifacts = project.getDependencyArtifacts();
        for (Artifact artifact : dependencyArtifacts ) {
            // Retrieve the dependency jar and its slsa file
            String artifactStr = artifact.getGroupId() + ":" + artifact.getArtifactId() + ":" + artifact.getVersion();
            try {
                // Retrieve the slsa file of the artifact
                executeMojo(
                    plugin(
                        groupId("com.googlecode.maven-download-plugin"),
                        artifactId("download-maven-plugin"),
                        version("1.7.0")
                    ),
                    goal("artifact"),
                    configuration(
                        element(name("outputDirectory"), "${project.build.directory}/slsa"),
                        element(name("groupId"), artifact.getGroupId()),
                        element(name("artifactId"), artifact.getArtifactId()),
                        element(name("version"), artifact.getVersion()),
                        element(name("type"), "intoto.build.slsa"),
                        element(name("classifier"), "jar")
                    ),
                    executionEnvironment(
                        project,
                        mavenSession,
                        pluginManager
                    )
                );

                // Retrieve the dependency jar if slsa file does exists for this artifact
                executeMojo(
                    plugin(
                        groupId("org.apache.maven.plugins"),
                        artifactId("maven-dependency-plugin"),
                        version("3.6.0")
                    ),
                    goal("copy"),
                    configuration(
                        element(name("outputDirectory"), "${project.build.directory}/slsa"),
                        element(name("artifact"), artifactStr)
                    ),
                    executionEnvironment(
                        project,
                        mavenSession,
                        pluginManager
                    )
                );
            } catch(MojoExecutionException e) {
                getLog().info("Skipping slsa verification for " + artifactStr + ": No slsa file found.");
                continue;
            }

            // Verify slsa file
            try {
                // Run slsa verification on the artifact and print the result
                // It will never fail the build process
                // This might be prone to command-injections. TODO: Secure against that. 
                String arguments = "verify-artifact --provenance-path ";
                arguments += "${project.build.directory}/slsa/" + artifact.getArtifactId() + "-" + artifact.getVersion() + "-jar.intoto.build.slsa ";
                arguments += " --source-uri ./ ${project.build.directory}/slsa/" + artifact.getArtifactId() + "-" + artifact.getVersion() + ".jar";
                executeMojo(
                    plugin(
                        groupId("org.codehaus.mojo"),
                        artifactId("exec-maven-plugin"),
                        version("3.1.0")
                    ),
                    goal("exec"),
                    configuration(
                        element(name("executable"), verifierPath),
                        element(name("commandlineArgs"), arguments),
                        element(name("useMavenLogger"), "true")
                    ),
                    executionEnvironment(
                        project,
                        mavenSession,
                        pluginManager
                    )
                );
            } catch(MojoExecutionException e) {
                // TODO: Properly interpret the output based on the verification plugin.
                getLog().info("Skipping slsa verification: Fail to run slsa verifier.");
                return;
            }
        }
    }
}