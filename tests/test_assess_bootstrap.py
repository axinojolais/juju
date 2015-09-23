from argparse import Namespace
from contextlib import contextmanager
from unittest import TestCase

from mock import patch

from assess_bootstrap import (
    assess_bootstrap,
    parse_args,
    )
from jujupy import (
    _temp_env as temp_env,
    )


class TestParseArgs(TestCase):

    def test_parse_args(self):
        args = parse_args(['foo', 'bar'])
        self.assertEqual(args, Namespace(
            juju='foo', env='bar', debug=False, region=None,
            temp_env_name=None))

    def test_parse_args_debug(self):
        args = parse_args(['foo', 'bar', '--debug'])
        self.assertEqual(args.debug, True)

    def test_parse_args_region(self):
        args = parse_args(['foo', 'bar', '--region', 'foo'])
        self.assertEqual(args.region, 'foo')

    def test_parse_args_temp_env_name(self):
        args = parse_args(['foo', 'bar', 'foo'])
        self.assertEqual(args.temp_env_name, 'foo')


class TestAssessBootstrap(TestCase):

    @contextmanager
    def assess_boostrap_cxt(self):
        call_cxt = patch('subprocess.call')
        cc_cxt = patch('subprocess.check_call')
        co_cxt = patch('subprocess.check_output', return_value='1.25.5')
        env_cxt = temp_env({'environments': {'bar': {'type': 'foo'}}})
        with call_cxt, cc_cxt, co_cxt, env_cxt:
            yield

    def test_assess_bootstrap_defaults(self):
        def check(myself):
            self.assertEqual(myself.env.config,
                             {'name': 'bar', 'type': 'foo'})
        with self.assess_boostrap_cxt():
            with patch('jujupy.EnvJujuClient.bootstrap', side_effect=check,
                       autospec=True):
                assess_bootstrap('/foo', 'bar', False, None, None)

    def test_assess_bootstrap_region_temp_env(self):
        def check(myself):
            self.assertEqual(
                myself.env.config, {
                    'name': 'qux', 'type': 'foo', 'region': 'baz'})
            self.assertEqual(myself.env.environment, 'qux')
        with self.assess_boostrap_cxt():
            with patch('jujupy.EnvJujuClient.bootstrap', side_effect=check,
                       autospec=True):
                assess_bootstrap('/foo', 'bar', False, 'baz', 'qux')
