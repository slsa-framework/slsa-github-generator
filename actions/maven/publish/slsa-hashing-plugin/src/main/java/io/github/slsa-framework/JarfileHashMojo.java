package dev.slsa.slsaframework;

import org.apache.maven.plugin.AbstractMojo;
import org.apache.maven.plugin.MojoExecutionException;
import org.apache.maven.plugin.MojoFailureException;
import org.apache.maven.plugins.annotations.LifecyclePhase;
import org.apache.maven.plugins.annotations.Mojo;
import org.apache.maven.plugins.annotations.Parameter;
import org.apache.maven.project.MavenProject;

import org.json.JSONObject;

import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.List;
import java.util.LinkedList;

@Mojo(name = "hash-jarfile", defaultPhase = LifecyclePhase.PACKAGE)
public class JarfileHashMojo extends AbstractMojo {
    private final String jsonBase = "{\"version\": 1, \"attestations\":[%ATTESTATIONS%]}";
    private final String attestationTemplate = "{\"name\": \"%NAME%\",\"subjects\":[{\"name\": \"%NAME%\",\"digest\":{\"sha256\":\"%HASH%\"}}]}";

    @Parameter(defaultValue = "${project}", required = true, readonly = true)
    private MavenProject project;

    @Parameter(property = "hash-jarfile.outputJsonPath", defaultValue = "")
    private String outputJsonPath;

    public void execute() throws MojoExecutionException, MojoFailureException {
        try {
            StringBuilder attestations = new StringBuilder();

            File targetDir = new File(project.getBasedir(), "target");
            File outputJson = this.getOutputJsonFile(targetDir.getAbsolutePath());
            for (File file : targetDir.listFiles()) {
                String filePath = file.getAbsolutePath();
                if (!filePath.endsWith("original") && (filePath.endsWith(".pom") || filePath.endsWith(".jar"))) {
                    byte[] data = Files.readAllBytes(file.toPath());
                    byte[] hash = MessageDigest.getInstance("SHA-256").digest(data);
                    String checksum = new BigInteger(1, hash).toString(16);

                    String attestation = attestationTemplate.replaceAll("%NAME%", file.getName());
                    attestation = attestation.replaceAll("%HASH%", checksum);
                    if (attestations.length() > 0) {
                        attestations.append(",");
                    }
                    attestations.append(attestation);
                }
            }
            String json = jsonBase.replaceAll("%ATTESTATIONS%", attestations.toString());

            Files.write(outputJson.toPath(), new JSONObject(json).toString(4).getBytes());
        } catch (IOException | NoSuchAlgorithmException e) {
            throw new MojoFailureException("Fail to generate hash for the jar files", e);
        }

    }

    private File getOutputJsonFile(String targetDir) {
        try {
            if (this.outputJsonPath != null && this.outputJsonPath.length() > 0) {
                File outputJson = new File(outputJsonPath);
                if (!outputJson.exists() || !outputJson.isFile()) {
                    outputJson.getParentFile().mkdirs();
                    Files.createFile(outputJson.toPath());
                }

                if (Files.isWritable(outputJson.toPath())) {
                    return outputJson;
                }
            }
            return new File(targetDir, "hash.json");
        } catch (IOException e) {
            return new File(targetDir, "hash.json");
        }
    }
}
