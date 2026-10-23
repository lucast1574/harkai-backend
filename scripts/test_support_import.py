import importlib.util
import pathlib
import unittest

spec = importlib.util.spec_from_file_location('registry', pathlib.Path(__file__).with_name('import-support-renipress.py'))
registry = importlib.util.module_from_spec(spec)
spec.loader.exec_module(registry)


class RegistryContactsTest(unittest.TestCase):
    def test_international_landline_and_separate_numbers(self):
        self.assertEqual(registry.phone('(51) 1 - 2190000', 'lima'), '+5112190000')
        self.assertEqual(registry.phones('219-0000 / 999888777', 'lima'), ['+5112190000', '+51999888777'])
        self.assertEqual(registry.phone('2190000 anexo 123', 'lima'), '')
        self.assertEqual(registry.phone('000-000000', 'lima'), '')

    def test_only_registry_image_links_are_retained(self):
        url = 'http://app20.susalud.gob.pe:8080/registro-renipress-webapp/ipress.htm?action=downloadFile&idArchivo=INSASU20211229105640.JPG'
        self.assertEqual(registry.photos({'IMAGEN_1': url}), [url])
        self.assertEqual(registry.photos({'IMAGEN_1': 'http://example.com/photo.jpg'}), [])


if __name__ == '__main__':
    unittest.main()
