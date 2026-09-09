import java.io.File;
import java.io.IOException;
import java.io.OutputStream;
import java.lang.reflect.Field;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import org.codehaus.modello.ModelloParameterConstants;
import org.codehaus.modello.core.DefaultGeneratorPluginManager;
import org.codehaus.modello.core.DefaultMetadataPluginManager;
import org.codehaus.modello.core.DefaultModelloCore;
import org.codehaus.modello.metadata.MetadataPlugin;
import org.codehaus.modello.model.Model;
import org.codehaus.modello.plugin.ModelloGenerator;
import org.codehaus.modello.plugin.java.JavaModelloGenerator;
import org.codehaus.modello.plugin.java.metadata.JavaMetadataPlugin;
import org.codehaus.modello.plugin.model.ModelMetadataPlugin;
import org.codehaus.modello.plugin.xpp3.Xpp3ReaderGenerator;
import org.codehaus.modello.plugin.xpp3.Xpp3WriterGenerator;
import org.codehaus.modello.plugins.xml.metadata.XmlMetadataPlugin;
import org.codehaus.plexus.build.BuildContext;
import org.codehaus.plexus.util.Scanner;

/** Plexus-free Modello driver for generating java/xpp3 sources from an .mdo. */
public final class RunModello {
	public static void main(String[] args) throws Exception {
		if (args.length < 3) {
			System.err.println("usage: RunModello <model.mdo> <version> <outputDir> [generator...]");
			System.exit(2);
		}
		File modelFile = new File(args[0]);
		String version = args[1];
		File outputDir = new File(args[2]);
		String[] generators = args.length > 3 ? java.util.Arrays.copyOfRange(args, 3, args.length)
				: new String[] {"java", "xpp3-reader", "xpp3-writer"};

		BuildContext buildContext = new NoopBuildContext();

		Map<String, MetadataPlugin> metadata = new LinkedHashMap<>();
		metadata.put("model", new ModelMetadataPlugin());
		metadata.put("java", new JavaMetadataPlugin());
		metadata.put("xml", new XmlMetadataPlugin());

		Map<String, ModelloGenerator> gens = new LinkedHashMap<>();
		JavaModelloGenerator javaGen = new JavaModelloGenerator();
		Xpp3ReaderGenerator xpp3Reader = new Xpp3ReaderGenerator();
		Xpp3WriterGenerator xpp3Writer = new Xpp3WriterGenerator();
		setField(javaGen, "buildContext", buildContext);
		setField(xpp3Reader, "buildContext", buildContext);
		setField(xpp3Writer, "buildContext", buildContext);
		gens.put("java", javaGen);
		gens.put("xpp3-reader", xpp3Reader);
		gens.put("xpp3-writer", xpp3Writer);

		DefaultMetadataPluginManager metadataManager = new DefaultMetadataPluginManager();
		setField(metadataManager, "plugins", metadata);
		DefaultGeneratorPluginManager generatorManager = new DefaultGeneratorPluginManager();
		setField(generatorManager, "plugins", gens);

		DefaultModelloCore core = new DefaultModelloCore();
		setField(core, "metadataPluginManager", metadataManager);
		setField(core, "generatorPluginManager", generatorManager);

		Model model = core.loadModel(modelFile);
		Map<String, Object> parameters = new HashMap<>();
		parameters.put(ModelloParameterConstants.OUTPUT_DIRECTORY, outputDir.getAbsolutePath());
		parameters.put(ModelloParameterConstants.VERSION, version);
		parameters.put(ModelloParameterConstants.PACKAGE_WITH_VERSION, "false");
		parameters.put(ModelloParameterConstants.OUTPUT_JAVA_SOURCE, "8");
		outputDir.mkdirs();
		for (String generator : generators) {
			core.generate(model, generator, parameters);
		}
	}

	static void setField(Object target, String name, Object value) throws Exception {
		Class<?> c = target.getClass();
		while (c != null) {
			try {
				Field f = c.getDeclaredField(name);
				f.setAccessible(true);
				f.set(target, value);
				return;
			} catch (NoSuchFieldException e) {
				c = c.getSuperclass();
			}
		}
		throw new NoSuchFieldException(target.getClass().getName() + "." + name);
	}

	static final class NoopBuildContext implements BuildContext {
		public boolean hasDelta(String relpath) {
			return true;
		}

		public boolean hasDelta(File file) {
			return true;
		}

		public boolean hasDelta(List relpaths) {
			return true;
		}

		public void refresh(File file) {}

		public OutputStream newFileOutputStream(File file) throws IOException {
			file.getParentFile().mkdirs();
			return java.nio.file.Files.newOutputStream(file.toPath());
		}

		public Scanner newScanner(File basedir) {
			return null;
		}

		public Scanner newDeleteScanner(File basedir) {
			return null;
		}

		public Scanner newScanner(File basedir, boolean ignoreDelta) {
			return null;
		}

		public boolean isIncremental() {
			return false;
		}

		public void setValue(String key, Object value) {}

		public Object getValue(String key) {
			return null;
		}

		public void addWarning(File file, int line, int column, String message, Throwable cause) {}

		public void addError(File file, int line, int column, String message, Throwable cause) {}

		public void addMessage(File file, int line, int column, String message, int severity, Throwable cause) {}

		public void removeMessages(File file) {}

		public boolean isUptodate(File target, File source) {
			return false;
		}
	}
}
