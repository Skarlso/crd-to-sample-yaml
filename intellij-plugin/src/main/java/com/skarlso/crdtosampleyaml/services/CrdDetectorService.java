package com.skarlso.crdtosampleyaml.services;

import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.vfs.VfsUtilCore;
import com.intellij.openapi.vfs.VirtualFile;
import org.yaml.snakeyaml.LoaderOptions;
import org.yaml.snakeyaml.Yaml;
import org.yaml.snakeyaml.constructor.SafeConstructor;

import java.io.StringReader;
import java.util.Map;

public class CrdDetectorService {

    private static final Logger LOG = Logger.getInstance(CrdDetectorService.class);
    private static final String CRD_KIND = "CustomResourceDefinition";
    private static final String CRD_API_GROUP = "apiextensions.k8s.io/";

    /**
     * Files above this are not worth parsing on every context menu update. A CRD large enough
     * to hit it is possible but rare, and the alternative is stalling the popup.
     */
    private static final long MAX_FILE_SIZE = 5L * 1024 * 1024;

    public static CrdDetectorService getInstance() {
        return new CrdDetectorService();
    }

    /**
     * snakeyaml's Yaml is not thread safe and detection now runs on background threads,
     * so each parse gets its own.
     */
    private static Yaml yaml() {
        return new Yaml(new SafeConstructor(new LoaderOptions()));
    }

    public boolean isCrdFile(VirtualFile file) {
        String content = readIfYaml(file);

        return content != null && isCrdContent(content);
    }

    public boolean isCrdContent(String content) {
        if (content == null || content.trim().isEmpty()) {
            return false;
        }

        try {
            // A CRD is often bundled with other manifests in one file.
            for (Object document : yaml().loadAll(new StringReader(content))) {
                if (isCrd(document)) {
                    return true;
                }
            }
        } catch (Exception e) {
            // Half-typed or templated YAML lands here constantly; it just isn't a CRD.
            LOG.debug("could not parse content as YAML", e);
        }

        return false;
    }

    public String extractCrdName(VirtualFile file) {
        String content = readIfYaml(file);
        if (content == null) {
            return null;
        }

        try {
            for (Object document : yaml().loadAll(new StringReader(content))) {
                if (!isCrd(document)) {
                    continue;
                }

                Object metadata = asMap(document).get("metadata");
                if (metadata instanceof Map) {
                    Object name = asMap(metadata).get("name");
                    if (name instanceof String) {
                        return (String) name;
                    }
                }
            }
        } catch (Exception e) {
            LOG.debug("could not extract CRD name from " + file.getPath(), e);
        }

        return null;
    }

    private boolean isCrd(Object document) {
        if (!(document instanceof Map)) {
            return false;
        }

        Map<String, Object> map = asMap(document);

        return CRD_KIND.equals(map.get("kind"))
                && map.get("apiVersion") instanceof String
                && ((String) map.get("apiVersion")).startsWith(CRD_API_GROUP);
    }

    @SuppressWarnings("unchecked")
    private Map<String, Object> asMap(Object document) {
        return (Map<String, Object>) document;
    }

    /**
     * Reads straight from the VFS rather than through the PSI, so this stays callable from a
     * background thread without holding a read action.
     */
    private String readIfYaml(VirtualFile file) {
        if (file == null || !file.isValid() || file.isDirectory() || !isYamlFile(file)) {
            return null;
        }

        if (file.getLength() > MAX_FILE_SIZE) {
            return null;
        }

        try {
            return VfsUtilCore.loadText(file);
        } catch (Exception e) {
            LOG.debug("could not read " + file.getPath(), e);

            return null;
        }
    }

    private boolean isYamlFile(VirtualFile file) {
        String extension = file.getExtension();

        return "yaml".equalsIgnoreCase(extension) || "yml".equalsIgnoreCase(extension);
    }
}
