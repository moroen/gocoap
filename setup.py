from setuptools import Extension
from setuptools import setup


setup(
    name='pycoap',
    description='A low level extension for COAP/COAPS-requests',
    url='https://github.com/moroen/python-coap-module',
    version='0.1.0',
    author='moroen',
    author_email='no@email.com',
    classifiers=[
        'License :: OSI Approved :: MIT License',
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