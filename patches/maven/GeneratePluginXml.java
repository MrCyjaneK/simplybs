package org.apache.maven.tools.plugin.generator;

import java.io.File;
import java.lang.reflect.Method;
import java.util.Collections;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

import org.apache.maven.artifact.Artifact;
import org.apache.maven.artifact.DefaultArtifact;
import org.apache.maven.artifact.handler.DefaultArtifactHandler;
import org.apache.maven.model.Build;
import org.apache.maven.plugin.descriptor.MojoDescriptor;
import org.apache.maven.plugin.descriptor.PluginDescriptor;
import org.apache.maven.project.MavenProject;
import org.apache.maven.tools.plugin.DefaultPluginToolsRequest;
import org.apache.maven.tools.plugin.extractor.annotations.JavaAnnotationsMojoDescriptorExtractor;
import org.apache.maven.tools.plugin.extractor.annotations.scanner.DefaultMojoAnnotationsScanner;
import org.apache.maven.tools.plugin.extractor.annotations.scanner.MojoAnnotatedClass;
import org.apache.maven.tools.plugin.extractor.annotations.scanner.MojoAnnotationsScannerRequest;

/**
 * Plexus-free driver: scan {@code @Mojo} classes and write {@code META-INF/maven/plugin.xml}.
 *
 * args: groupId artifactId version goalPrefix classesDir outputDir [scanClasspath]
 */
public final class GeneratePluginXml {
	public static void main(String[] args) throws Exception {
		if (args.length < 6) {
			System.err.println(
					"usage: GeneratePluginXml <groupId> <artifactId> <version> <goalPrefix> <classesDir> <outputDir> [scanClasspath]");
			System.exit(2);
		}
		String groupId = args[0];
		String artifactId = args[1];
		String version = args[2];
		String goalPrefix = args[3];
		File classesDir = new File(args[4]).getAbsoluteFile();
		File outputDir = new File(args[5]).getAbsoluteFile();
		String scanClasspath = args.length > 6 ? args[6] : "";

		DefaultArtifactHandler handler = new DefaultArtifactHandler("jar");
		Artifact projectArtifact = new DefaultArtifact(groupId, artifactId, version, "compile", "jar", null, handler);

		MavenProject project = new MavenProject();
		project.setGroupId(groupId);
		project.setArtifactId(artifactId);
		project.setVersion(version);
		project.setName(artifactId);
		project.setArtifact(projectArtifact);
		Build build = new Build();
		build.setOutputDirectory(classesDir.getPath());
		build.setDirectory(classesDir.getParentFile() != null ? classesDir.getParentFile().getPath() : classesDir.getPath());
		project.setBuild(build);

		PluginDescriptor pluginDescriptor = new PluginDescriptor();
		pluginDescriptor.setGroupId(groupId);
		pluginDescriptor.setArtifactId(artifactId);
		pluginDescriptor.setVersion(version);
		pluginDescriptor.setGoalPrefix(goalPrefix);
		pluginDescriptor.setName(artifactId);
		pluginDescriptor.setDependencies(Collections.emptyList());

		Set<Artifact> deps = new LinkedHashSet<Artifact>();
		for (String path : scanClasspath.split(":")) {
			if (path.isEmpty()) {
				continue;
			}
			File f = new File(path);
			if (!f.exists() || f.equals(classesDir)) {
				continue;
			}
			deps.add(artifactFor(f));
		}

		MojoAnnotationsScannerRequest scanReq = new MojoAnnotationsScannerRequest();
		scanReq.setClassesDirectories(Collections.singletonList(classesDir));
		scanReq.setDependencies(deps);
		scanReq.setProject(project);

		DefaultMojoAnnotationsScanner scanner = new DefaultMojoAnnotationsScanner();
		Map<String, MojoAnnotatedClass> scanned = scanner.scan(scanReq);

		JavaAnnotationsMojoDescriptorExtractor extractor = new JavaAnnotationsMojoDescriptorExtractor();
		Method toMojos = JavaAnnotationsMojoDescriptorExtractor.class.getDeclaredMethod(
				"toMojoDescriptors", Map.class, PluginDescriptor.class);
		toMojos.setAccessible(true);
		@SuppressWarnings("unchecked")
		List<MojoDescriptor> mojos = (List<MojoDescriptor>) toMojos.invoke(extractor, scanned, pluginDescriptor);
		if (mojos == null || mojos.isEmpty()) {
			System.err.println("GeneratePluginXml: no @Mojo classes in " + classesDir);
			System.exit(1);
		}
		for (MojoDescriptor mojo : mojos) {
			pluginDescriptor.addMojo(mojo);
		}

		DefaultPluginToolsRequest request = new DefaultPluginToolsRequest(project, pluginDescriptor);
		request.setEncoding("UTF-8");
		request.setDependencies(deps);

		outputDir.mkdirs();
		PluginDescriptorFilesGenerator generator = new PluginDescriptorFilesGenerator();
		generator.writeDescriptor(
				new File(outputDir, "plugin.xml"), request, PluginDescriptorFilesGenerator.DescriptorType.STANDARD);
		System.out.println("wrote " + new File(outputDir, "plugin.xml") + " (" + mojos.size() + " mojos)");
	}

	private static Artifact artifactFor(File file) {
		String name = file.getName();
		String groupId = "scan";
		String artifactId = name;
		String version = "1";
		if (name.endsWith(".jar")) {
			artifactId = name.substring(0, name.length() - 4);
		}
		if (name.startsWith("maven-plugin-api-")) {
			groupId = "org.apache.maven";
			artifactId = "maven-plugin-api";
			version = name.substring("maven-plugin-api-".length()).replace(".jar", "");
		} else if (name.startsWith("maven-api-core-")) {
			groupId = "org.apache.maven";
			artifactId = "maven-api-core";
			version = name.substring("maven-api-core-".length()).replace(".jar", "");
		}
		DefaultArtifactHandler ah = new DefaultArtifactHandler("jar");
		DefaultArtifact artifact = new DefaultArtifact(groupId, artifactId, version, "compile", "jar", null, ah);
		artifact.setFile(file);
		return artifact;
	}
}
