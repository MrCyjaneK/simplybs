import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;
import java.util.jar.JarOutputStream;

/** Relocate Java package names inside jars (constant-pool UTF-8 + entry paths). */
public final class RelocateJar {
	public static void main(String[] args) throws IOException {
		String out = null;
		List<String> inputs = new ArrayList<>();
		List<String[]> rules = new ArrayList<>();
		for (int i = 0; i < args.length; i++) {
			if ("-o".equals(args[i])) {
				out = args[++i];
			} else if ("--".equals(args[i])) {
				for (int j = i + 1; j < args.length; j += 2) {
					rules.add(new String[] {args[j], args[j + 1]});
				}
				break;
			} else {
				inputs.add(args[i]);
			}
		}
		if (out == null || inputs.isEmpty() || rules.isEmpty()) {
			System.err.println(
					"usage: RelocateJar -o out.jar in.jar [in.jar ...] -- oldpkg newpkg ...");
			System.exit(2);
		}
		Map<String, byte[]> entries = new LinkedHashMap<>();
		for (String in : inputs) {
			try (JarFile jf = new JarFile(in)) {
				jf.stream().forEach(e -> {
					if (e.isDirectory()) {
						return;
					}
					String name = rewriteString(e.getName(), rules);
					if (name.endsWith("module-info.class") || name.endsWith("META-INF/MANIFEST.MF")) {
						return;
					}
					try (InputStream is = jf.getInputStream(e)) {
						byte[] data = is.readAllBytes();
						if (name.endsWith(".class")) {
							data = rewriteClass(data, rules);
						}
						entries.put(name, data);
					} catch (IOException ex) {
						throw new RuntimeException(ex);
					}
				});
			}
		}
		Path outPath = Path.of(out);
		Files.createDirectories(outPath.getParent());
		try (JarOutputStream jos = new JarOutputStream(Files.newOutputStream(outPath))) {
			for (Map.Entry<String, byte[]> e : entries.entrySet()) {
				jos.putNextEntry(new JarEntry(e.getKey()));
				jos.write(e.getValue());
				jos.closeEntry();
			}
		}
	}

	static String rewriteString(String s, List<String[]> rules) {
		for (String[] r : rules) {
			s = s.replace(r[0].replace('.', '/'), r[1].replace('.', '/'));
			s = s.replace(r[0], r[1]);
		}
		return s;
	}

	static byte[] rewriteClass(byte[] data, List<String[]> rules) {
		if (data.length < 10 || data[0] != (byte) 0xca || data[1] != (byte) 0xfe || data[2] != (byte) 0xba
				|| data[3] != (byte) 0xbe) {
			return data;
		}
		int count = u2(data, 8);
		ByteArrayOutputStream out = new ByteArrayOutputStream();
		out.write(data, 0, 10);
		int i = 10;
		int n = 1;
		while (n < count) {
			int tag = data[i] & 0xff;
			out.write(tag);
			i++;
			if (tag == 1) {
				int length = u2(data, i);
				i += 2;
				byte[] raw = java.util.Arrays.copyOfRange(data, i, i + length);
				i += length;
				String text = new String(raw, StandardCharsets.UTF_8);
				String neu = rewriteString(text, rules);
				byte[] encoded = neu.getBytes(StandardCharsets.UTF_8);
				out.write((encoded.length >> 8) & 0xff);
				out.write(encoded.length & 0xff);
				out.write(encoded, 0, encoded.length);
			} else if (tag == 7 || tag == 8 || tag == 16 || tag == 19 || tag == 20) {
				out.write(data, i, 2);
				i += 2;
			} else if (tag == 3 || tag == 4 || tag == 9 || tag == 10 || tag == 11 || tag == 12 || tag == 17
					|| tag == 18) {
				out.write(data, i, 4);
				i += 4;
			} else if (tag == 5 || tag == 6) {
				out.write(data, i, 8);
				i += 8;
				n++;
			} else if (tag == 15) {
				out.write(data, i, 3);
				i += 3;
			} else {
				throw new IllegalStateException("unknown constant tag " + tag);
			}
			n++;
		}
		out.write(data, i, data.length - i);
		return out.toByteArray();
	}

	static int u2(byte[] data, int i) {
		return ((data[i] & 0xff) << 8) | (data[i + 1] & 0xff);
	}
}
