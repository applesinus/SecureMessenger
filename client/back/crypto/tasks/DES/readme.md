# 1 - DES/DEAL

<a name="Russian"></a>
## Русский
[Jump to English](#English)

Все задания выполняются на объектно-ориентированном языке программирования. Применение готовых реализаций алгоритмов защиты  информации и библиотек, содержащих такие реализации, не допускается.

1. Реализуйте функцию для выполнения перестановки битов в рамках переданного значения (тип значения - массив байтов). Параметры функции: значение для перестановки, правило перестановки (P-блок), правила индексирования битов: биты индексируются от младшего к старшему или наоборот; номер начального бита == 0 или == 1.

2. Спроектируйте следующие сущности:

    1. интерфейс, предоставляющий описание функционала для процедуры расширения ключа (генерации раундовых ключей) (параметр метода: входной ключ - массив байтов, результат - массив раундовых ключей (каждый раундовый ключ - массив байтов));

    2. интерфейс, предоставляющий описание функционала по выполнению шифрующего преобразования (параметры метода: входной блок - массив байтов, раундовый ключ - массив байтов, результат: выходной блок - массив байтов);

    3. интерфейс, предоставляющий описание функционала по выполнению шифрования и дешифрования симметричным алгоритмом (параметр методов: [де]шифруемый блок (массив байтов)) с преднастроенными отдельным методом раундовыми ключами (параметр метода: ключ [де]шифрования (массив байтов));

    4. класс, репрезентирующий контекст выполнения симметричного криптографического алгоритма, предоставляющий объектный функционал по выполнению операций шифрования и дешифрования заданным ключом симметричного алгоритма (реализацией интерфейса из п. 3) с поддержкой одного из режимов шифрования (задаётся перечислением): **ECB, CBC, PCBC, CFB, OFB, CTR, Random Delta**; а также с поддержкой одного из режимов набивки (задаётся перечислением): **Zeros, ANSI X.923, PKCS7, ISO 10126**. Параметры конструктора объекта класса: ключ шифрования, режим шифрования (объект перечисления), режим набивки (объект перечисления), вектор инициализации для заданного режима (опционально), дополнительные параметры для указанного режима (коллекция аргументов переменной длины). Параметры перегруженных методов шифрования/дешифрования: данные для [де]шифрования (массив байтов произвольной длины) и ссылка на результирующий массив байтов, либо путь к файлу со входными данными и путь к файлу с результатом [де]шифрования). Где возможно, реализуйте распараллеливание вычислений. Выполнение операций шифрования/дешифрования должно производиться асинхронно.

3. На базе интерфейса 2 из задания спроектируйте и реализуйте класс, реализующий функционал сети Фейстеля. Конструктор класса должен принимать в качестве параметров реализации интерфейсов 2.1 и 2.2.

4. Реализуйте алгоритм шифрования DES на базе класса из задания 3, определив свои реализации интерфейсов 2.1 и 2.2. При реализации DES используйте функцию, реализованную в задании 1.

5. Продемонстрируйте выполнение шифрования и дешифрования псевдослучайных последовательностей байтов и файлов (текстовых, музыкальных, изображений, видео, файлов исходного кода тестов на аллокатор на базе красно-чёрного дерева и т. д.) алгоритмом DES с использованием различных режимов шифрования при помощи типов, реализованных в заданиях 2-4.

6. Реализуйте алгоритм шифрования DEAL на базе класса из задания 3, определив свои реализации интерфейсов 2.1 и 2.2. Для внедрения Вашей реализации алгоритма DES в алгоритм DEAL реализуйте адаптер, позволяющий использовать реализации алгоритма DES в качестве раундовой функции F.

7. Продемонстрируйте выполнение шифрования и дешифрования псевдослучайных последовательностей байтов и файлов (текстовых, музыкальных, изображений, видео, файлов исходного кода тестов на умножение Шёнхаге-Штрассена и т. д.) алгоритмом DEAL с использованием различных режимов шифрования при помощи типов, реализованных в заданиях 2-6.

<a name="English"></a>
## English
[Перейти к русскому](#Russian)

(translated by AI)

All tasks are to be completed in an object-oriented programming language. The use of ready-made implementations of information protection algorithms and libraries containing such implementations is not allowed.

1. Implement a function to perform bit permutation within a given value (value type - byte array). Function parameters: value for permutation, permutation rule (P-box), bit indexing rules: bits are indexed from least significant to most significant or vice versa; the starting bit number is either 0 or 1.

2. Design the following entities:

    1. An interface providing the functionality description for the key expansion procedure (generating round keys) (method parameter: input key - byte array, result - array of round keys (each round key - byte array));

    2. An interface providing the functionality description for performing an encryption transformation (method parameters: input block - byte array, round key - byte array, result: output block - byte array);

    3. An interface providing the functionality description for performing encryption and decryption with a symmetric algorithm (method parameters: [de]cipher block (byte array)) with round keys preconfigured by a separate method (method parameter: [de]cipher key (byte array));

    4. A class representing the context of executing a symmetric cryptographic algorithm, providing object-oriented functionality for performing encryption and decryption operations with a given symmetric algorithm key (implementing the interface from point 3) with support for one of the encryption modes (specified by an enumeration): **ECB, CBC, PCBC, CFB, OFB, CTR, Random Delta**; as well as support for one of the padding modes (specified by an enumeration): **Zeros, ANSI X.923, PKCS7, ISO 10126**. Class constructor parameters: encryption key, encryption mode (enumeration object), padding mode (enumeration object), initialization vector for the specified mode (optional), additional parameters for the specified mode (variable-length argument collection). Parameters of the overloaded encryption/decryption methods: data to be [de]ciphered (byte array of arbitrary length) and a reference to the resulting byte array, or the path to the input data file and the path to the [de]crypted result file. Where possible, implement parallelization of computations. The execution of encryption/decryption operations should be asynchronous.

3. Based on interface 2 from the task, design and implement a class that implements the functionality of the Feistel network. The class constructor should take as parameters the implementations of interfaces 2.1 and 2.2.

4. Implement the DES encryption algorithm based on the class from task 3, defining your own implementations of interfaces 2.1 and 2.2. When implementing DES, use the function implemented in task 1.

5. Demonstrate the execution of encryption and decryption of pseudorandom byte sequences and files (text, music, images, video, source code files of tests on the allocator based on a red-black tree, etc.) using the DES algorithm with different encryption modes using the types implemented in tasks 2-4.

6. Implement the DEAL encryption algorithm based on the class from task 3, defining your own implementations of interfaces 2.1 and 2.2. To integrate your DES implementation into the DEAL algorithm, implement an adapter that allows the use of the DES algorithm implementations as the round function F.

7. Demonstrate the execution of encryption and decryption of pseudorandom byte sequences and files (text, music, images, video, source code files of tests on the Schönhage-Strassen multiplication, etc.) using the DEAL algorithm with different encryption modes using the types implemented in tasks 2-6.