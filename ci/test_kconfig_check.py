"""Tests fuer ci/kconfig_check.py. Laufen im Gate ci-selftest."""

import io
import os
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.dirname(__file__))
import kconfig_check  # noqa: E402


def write(directory, name, text):
    path = os.path.join(directory, name)
    with open(path, "w", encoding="utf-8") as handle:
        handle.write(text)
    return path


class Check(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.dir = self.tmp.name
        self.required = write(self.dir, "required.config",
                              "# Grund: USB-HID zum Zielrechner\nCONFIG_USB_CONFIGFS_F_HID=y\n"
                              "CONFIG_I2C_GPIO=m\n")
        self.remove = write(self.dir, "remove.config",
                            "# Grund: kein Bluetooth\n# CONFIG_BT is not set\n")
        self.harden = write(self.dir, "harden.config",
                            "# Grund: Usercopy-Pruefung\nCONFIG_HARDENED_USERCOPY=y\n")

    def tearDown(self):
        self.tmp.cleanup()

    def run_check(self, config_text, strict=False):
        config = write(self.dir, ".config", config_text)
        argv = ["kconfig_check.py", config, "--required", self.required,
                "--remove", self.remove, "--harden", self.harden]
        if strict:
            argv.append("--strict")
        out = io.StringIO()
        old = sys.stdout
        sys.stdout = out
        try:
            rc = kconfig_check.main(argv)
        finally:
            sys.stdout = old
        return rc, out.getvalue()

    def test_missing_required_option_fails(self):
        rc, out = self.run_check("# CONFIG_USB_CONFIGFS_F_HID is not set\nCONFIG_I2C_GPIO=m\n")
        self.assertEqual(rc, 1)
        self.assertIn("CONFIG_USB_CONFIGFS_F_HID", out)

    def test_required_module_counts_as_built_in_or_module(self):
        rc, _ = self.run_check("CONFIG_USB_CONFIGFS_F_HID=y\nCONFIG_I2C_GPIO=y\n")
        self.assertEqual(rc, 0)

    def test_option_absent_from_config_is_not_set(self):
        rc, out = self.run_check("CONFIG_I2C_GPIO=m\n")
        self.assertEqual(rc, 1)
        self.assertIn("CONFIG_USB_CONFIGFS_F_HID", out)

    def test_leftovers_are_reported_but_only_fail_in_strict_mode(self):
        config = "CONFIG_USB_CONFIGFS_F_HID=y\nCONFIG_I2C_GPIO=m\nCONFIG_BT=y\n"
        rc, out = self.run_check(config)
        self.assertEqual(rc, 0)
        self.assertIn("CONFIG_BT", out)
        self.assertIn("CONFIG_HARDENED_USERCOPY", out)
        rc, _ = self.run_check(config, strict=True)
        self.assertEqual(rc, 1)

    def test_fully_trimmed_config_passes_strict(self):
        rc, _ = self.run_check("CONFIG_USB_CONFIGFS_F_HID=y\nCONFIG_I2C_GPIO=m\n"
                               "CONFIG_HARDENED_USERCOPY=y\n", strict=True)
        self.assertEqual(rc, 0)

    def test_fragment_entry_without_reason_is_rejected(self):
        write(self.dir, "remove.config", "# CONFIG_BT is not set\n")
        rc, out = self.run_check("CONFIG_USB_CONFIGFS_F_HID=y\nCONFIG_I2C_GPIO=m\n")
        self.assertEqual(rc, 2)
        self.assertIn("Grund", out)

    def test_symbol_in_two_fragments_is_rejected(self):
        write(self.dir, "harden.config", "# Grund: x\n# CONFIG_BT is not set\n")
        rc, _ = self.run_check("CONFIG_USB_CONFIGFS_F_HID=y\nCONFIG_I2C_GPIO=m\n")
        self.assertEqual(rc, 2)


class Fragments(unittest.TestCase):
    """Die echten Fragmente des Forks muessen in sich stimmig sein."""

    def test_repository_fragments_are_consistent(self):
        base = os.path.join(os.path.dirname(__file__), "kernel")
        fragments = [os.path.join(base, f) for f in
                     ("required.config", "remove.config", "harden.config")]
        entries = kconfig_check.load_fragments(fragments)
        self.assertGreater(len(entries), 40)


if __name__ == "__main__":
    unittest.main()
