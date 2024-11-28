# 2 - RSA

<a name="Russian"></a>
## Русский
[Jump to English](#English)

Все задания выполняются на объектно-ориентированном языке программирования. Применение готовых реализаций алгоритмов защиты информации и библиотек, содержащих такие реализации, не допускается. Допускается использование существующих (входящих в ядро и сторонних) реализаций длинных целых и длинных вещественных чисел.

1. Спроектируйте и реализуйте stateless-сервис с компонентным функционалом для:
    - вычисления символа Лежандра;
    - вычисления символа Якоби;
    - вычисления НОД двух целых чисел при помощи алгоритма Евклида;
    - вычисления НОД двух целых чисел и решения соотношения Безу при помощи расширенного алгоритма Евклида;
    - выполнения операции возведения в степень по модулю.
Для реализованного сервиса не допускается адаптирование функционала, предоставляемого ядром используемого ЯП и сторонними библиотеками. Продемонстрируйте работу реализованного функционала.

2. Спроектируйте интерфейс, предоставляющий описание функционала для вероятностного теста простоты (параметры метода: тестируемое значение, минимальная вероятность простоты в диапазоне [0.5, 1) ). На базе спроектированного интерфейса и поведенческого паттерна проектирования “Шаблонный метод” реализуйте базовый абстрактный класс для вероятностного теста простоты, с возможностью кастомизации поведения одной итерации теста. С использованием сервиса, реализованного в задании 1, пронаследуйте базовый класс для реализации следующих вероятностных тестов простоты: Ферма, Соловея-Штрассена, Миллера-Рабина.

3. Спроектируйте и реализуйте объектный сервис, предназначенный для выполнения шифрования и дешифрования данных алгоритмом RSA. Сервис должен содержать объект вложенного (nested) сервиса для генерации ключей алгоритма RSA (контракт конструктора вложенного сервиса: используемый тест простоты (задаётся перечислением, тип которого является nested по отношению к типу сервиса для выполнения шифрования/дешифрования алгоритмом RSA), минимальная вероятность простоты в диапазоне [0.5, 1), битовая длина генерируемых проверяемых выбранным тестом простоты псевдослучайных чисел; параметры делегируются из конструктора сервиса-обёртки). При генерации ключей обеспечьте невозможность применимости атаки Ферма и атаки Винера. Новую ключевую пару можно генерировать произвольное количество раз. Продемонстрируйте выполнение шифрования и дешифрования данных алгоритмом RSA посредством реализованных сервисов.

4. Реализуйте сервис, предоставляющий компонентный функционал для выполнения атаки Ферма на открытый ключ алгоритма RSA. Для данного открытого ключа в качестве результата выполнения атаки необходимо получить найденное значение дешифрующей экспоненты, а также значение функции Эйлера от модуля RSA.

5. Реализуйте сервис, предоставляющий компонентный функционал для выполнения атаки Винера на открытый ключ алгоритма RSA. Для данного открытого ключа в качестве результата выполнения атаки необходимо получить найденное значение дешифрующей экспоненты, значение функции Эйлера от модуля RSA, а также коллекцию вычисленных во время атаки подходящих дробей для дроби, построенной из компонентов открытого ключа.

<a name="English"></a>
## English
[Перейти к русскому](#Russian)

(translated by AI)

All tasks are to be completed in an object-oriented programming language. The use of ready-made implementations of information protection algorithms and libraries containing such implementations is not allowed.

1. Design and implement a stateless service with component functionality for:
    - calculating the Legendre symbol;
    - calculating the Jacobi symbol;
    - calculating the GCD of two integers using the Euclidean algorithm;
    - calculating the GCD of two integers and solving Bézout's identity using the extended Euclidean algorithm;
    - performing modular exponentiation.
For the implemented service, adapting functionality provided by the core of the used programming language and third-party libraries is not allowed. Demonstrate the functionality of the implemented service.

2. Design an interface that provides the functionality description for a probabilistic primality test (method parameters: value to be tested, minimum probability of primality in the range [0.5, 1)). Based on the designed interface and the behavioral design pattern "Template Method," implement a basic abstract class for a probabilistic primality test, with the ability to customize the behavior of one test iteration. Using the service implemented in task 1, inherit the base class to implement the following probabilistic primality tests: Fermat, Solovay-Strassen, Miller-Rabin.

3. Design and implement an object-oriented service for performing encryption and decryption of data using the RSA algorithm. The service should contain a nested service object for generating RSA keys (nested service constructor contract: primality test used (specified by an enumeration, the type of which is nested relative to the type of the service for performing RSA encryption/decryption), minimum probability of primality in the range [0.5, 1), bit length of pseudorandom numbers generated by the selected primality test; parameters are delegated from the wrapper service constructor). Ensure that the generated keys are not susceptible to Fermat's attack and Wiener's attack. A new key pair can be generated an arbitrary number of times. Demonstrate the execution of encryption and decryption of data using the RSA algorithm through the implemented services.

4. Implement a service that provides component functionality for performing Fermat's attack on the public key of the RSA algorithm. For the given public key, as a result of the attack, you need to obtain the found value of the decrypting exponent, as well as the value of Euler's function of the RSA modulus.

5. Implement a service that provides component functionality for performing Wiener's attack on the public key of the RSA algorithm. For the given public key, as a result of the attack, you need to obtain the found value of the decrypting exponent, the value of Euler's function of the RSA modulus, as well as a collection of calculated during the attack suitable fractions for the fraction constructed from the components of the public key.