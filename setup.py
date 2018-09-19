from setuptools import Extension
from setuptools import setup


setup(
    name='tradfri-coap',
    description='Examples for https://github.com/asottile/setuptools-golang',
    url='https://github.com/asottile/setuptools-golang-examples',
    version='0.2.0',
    author='Moroen',
    author_email='moroen@gmail.com',
    classifiers=[
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python :: 2',
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: Implementation :: CPython',
        'Programming Language :: Python :: Implementation :: PyPy',
    ],
    ext_modules=[
        Extension('pycoap', ['coap/coap.go']),
    ],
    build_golang={'root': 'github.com/moroen/python-coap-module'},
    setup_requires=['setuptools-golang>=0.2.0'],
)