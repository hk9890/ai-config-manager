package repo

import (
	"io"
	"os"
	"path/filepath"
)

func (m *Manager) copyFile(src, dst string) error {
	// Get source file info for logging
	srcInfo, err := os.Stat(src)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to stat source file",
				"source", src,
				"error", err.Error(),
			)
		}
		return err
	}

	if m.logger != nil {
		m.logger.Debug("copying file",
			"source", src,
			"dest", dst,
			"size", srcInfo.Size(),
			"permissions", srcInfo.Mode(),
		)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to open source file",
				"source", src,
				"error", err.Error(),
			)
		}
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to create destination file",
				"dest", dst,
				"error", err.Error(),
			)
		}
		return err
	}
	defer destFile.Close()

	bytesWritten, err := io.Copy(destFile, sourceFile)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to copy file contents",
				"source", src,
				"dest", dst,
				"bytes_written", bytesWritten,
				"error", err.Error(),
			)
		}
		return err
	}

	if err := destFile.Sync(); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to sync destination file",
				"dest", dst,
				"error", err.Error(),
			)
		}
		return err
	}

	if m.logger != nil {
		m.logger.Debug("file copied successfully",
			"source", src,
			"dest", dst,
			"bytes", bytesWritten,
		)
	}

	return nil
}

// copyDir recursively copies a directory from src to dst
func (m *Manager) copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to stat source directory",
				"source", src,
				"error", err.Error(),
			)
		}
		return err
	}

	if m.logger != nil {
		m.logger.Debug("copying directory",
			"source", src,
			"dest", dst,
			"permissions", srcInfo.Mode(),
		)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to create destination directory",
				"dest", dst,
				"permissions", srcInfo.Mode(),
				"error", err.Error(),
			)
		}
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to read source directory",
				"source", src,
				"error", err.Error(),
			)
		}
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Follow symlinks with os.Stat
		info, err := os.Stat(srcPath)
		if err != nil {
			// Skip entries we can't stat
			if m.logger != nil {
				m.logger.Debug("skipping entry that cannot be stat'd",
					"path", srcPath,
					"error", err.Error(),
				)
			}
			continue
		}

		if info.IsDir() {
			// Recursively copy subdirectory
			if err := m.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := m.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	if m.logger != nil {
		m.logger.Debug("directory copied successfully",
			"source", src,
			"dest", dst,
			"entries", len(entries),
		)
	}

	return nil
}

// BulkImportOptions contains options for bulk import operations
