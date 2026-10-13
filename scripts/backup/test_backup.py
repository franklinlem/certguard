"""Regressões de segurança do backup, com transporte/disco simulados e GPG real."""
import fcntl
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).with_name('backup.sh').resolve()
RESTORE = Path(__file__).with_name('restore-test.sh').resolve()

@unittest.skipUnless(shutil.which('gpg'), 'GnuPG obrigatório para ensaio de cifragem')
class BackupTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        self.home = self.root / 'home'
        self.state = self.home / '.local/share/certguard-backup'
        self.keyring = self.state / 'gnupg'
        self.keyring.mkdir(parents=True, mode=0o700)
        self.mount = self.root / 'storage'
        self.backups = self.mount / 'backups/certguard'
        self.backups.mkdir(parents=True, mode=0o700)
        self.bin = self.root / 'bin'
        self.bin.mkdir()
        self.log = self.root / 'ssh-called'
        self.env = dict(os.environ, HOME=str(self.home), CERTGUARD_BACKUP_MOUNT=str(self.mount),
                        PATH=str(self.bin) + os.pathsep + os.environ['PATH'])
        self.mock('hostname', 'printf frank-note2')
        self.mock('mountpoint', 'exit 0')
        self.mock('findmnt', 'printf disk-test')
        self.mock('df', "printf 'Avail\\n999999999\\n'")
        self.mock('ssh', f"touch '{self.log}'; printf PGDMP-test-fixture")
        (self.state / 'storage.uuid').write_text('disk-test\n')
        subprocess.run(['gpg', '--homedir', str(self.keyring), '--batch', '--pinentry-mode', 'loopback',
                        '--passphrase', '', '--quick-generate-key', 'CertGuard disposable test',
                        'default', 'default', '1d'], check=True, capture_output=True)
        keys = subprocess.check_output(['gpg', '--homedir', str(self.keyring), '--with-colons', '--list-keys'], stderr=subprocess.DEVNULL).decode()
        self.fingerprint = next(s.split(':')[9] for s in keys.splitlines() if s.startswith('fpr:'))
        (self.state / 'recipient').write_text(self.fingerprint + '\n')
        self.sentinel = self.backups / 'existing-backup.txt'
        self.sentinel.write_text('preservar')

    def tearDown(self):
        subprocess.run(['gpgconf', '--homedir', str(self.keyring), '--kill', 'gpg-agent'], capture_output=True)
        self.temp.cleanup()

    def mock(self, name, body):
        p = self.bin / name
        p.write_text('#!/bin/sh\n' + body + '\n')
        p.chmod(0o755)

    def run_backup(self, environment='homologacao'):
        return subprocess.run(['bash', str(SCRIPT), environment], env=self.env, capture_output=True)

    def assert_no_backup(self):
        self.assertEqual([], list(self.backups.rglob('*.dump.gpg')))
        self.assertEqual([], list(self.backups.rglob('.certguard-*')))
        self.assertEqual('preservar', self.sentinel.read_text())

    def test_encrypts_and_preserves_existing_files(self):
        for _ in range(2):
            result = self.run_backup()
            self.assertEqual(0, result.returncode, result.stderr.decode())
        files = list(self.backups.rglob('*.dump.gpg'))
        self.assertEqual(2, len(files))
        for file in files:
            self.assertNotIn(b'PGDMP-test-fixture', file.read_bytes())
            plaintext = subprocess.check_output(['gpg', '--homedir', str(self.keyring), '--batch', '--decrypt', str(file)], stderr=subprocess.DEVNULL)
            self.assertEqual(b'PGDMP-test-fixture', plaintext)
            subprocess.run(['sha256sum', '--check', '--status', file.name + '.sha256'], cwd=file.parent, check=True)
        self.assertEqual('preservar', self.sentinel.read_text())
        self.assertEqual([], list(self.backups.rglob('.certguard-*')))

    def test_missing_or_wrong_disk_stops_before_ssh(self):
        for body in ['exit 1', 'exit 0']:
            self.mock('mountpoint', body)
            self.mock('findmnt', 'printf wrong-disk')
            self.assertNotEqual(0, self.run_backup().returncode)
            self.assertFalse(self.log.exists())
            self.assert_no_backup()

    def test_invalid_environment_and_unconfigured_production(self):
        for environment in ['../outro', 'producao']:
            self.assertNotEqual(0, self.run_backup(environment).returncode)
            self.assertFalse(self.log.exists())
            self.assert_no_backup()

    def test_failed_ssh_does_not_publish_partial_backup(self):
        self.mock('ssh', 'printf partial-fixture; exit 7')
        self.assertNotEqual(0, self.run_backup().returncode)
        self.assert_no_backup()

    def test_failed_encryption_does_not_publish(self):
        self.mock('gpg', 'exit 9')
        self.assertNotEqual(0, self.run_backup().returncode)
        self.assert_no_backup()

    def test_restore_rejects_tampering_before_creating_container(self):
        self.assertEqual(0, self.run_backup().returncode)
        file = next(self.backups.rglob('*.dump.gpg'))
        docker_log = self.root / 'docker-called'
        self.mock('docker', f"touch '{docker_log}'; exit 11")
        file.write_bytes(file.read_bytes() + b'tampered')
        result = subprocess.run(['bash', str(RESTORE), str(file)], env=self.env, capture_output=True)
        self.assertNotEqual(0, result.returncode)
        self.assertFalse(docker_log.exists())
        self.assertTrue(file.exists())
        self.assertEqual('preservar', self.sentinel.read_text())

    def test_restore_rejects_file_outside_backup_directory(self):
        docker_log = self.root / 'docker-called'
        self.mock('docker', f"touch '{docker_log}'; exit 11")
        outside = self.root / 'certguard-fixture.dump.gpg'
        outside.write_bytes(b'fixture')
        result = subprocess.run(['bash', str(RESTORE), str(outside)], env=self.env, capture_output=True)
        self.assertNotEqual(0, result.returncode)
        self.assertFalse(docker_log.exists())
        self.assertTrue(outside.exists())

    def test_lock_prevents_concurrent_backup(self):
        lock = self.home / '.local/state/certguard-backup/backup.lock'
        lock.parent.mkdir(parents=True)
        with lock.open('w') as stream:
            fcntl.flock(stream, fcntl.LOCK_EX | fcntl.LOCK_NB)
            self.assertNotEqual(0, self.run_backup().returncode)
            self.assertFalse(self.log.exists())
            self.assert_no_backup()

if __name__ == '__main__':
    unittest.main()
