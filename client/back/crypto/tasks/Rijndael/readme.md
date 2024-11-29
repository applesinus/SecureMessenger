# 3 - Rijndael

<a name="Russian"></a>
## Русский
[Jump to English](#English)

Все задания выполняются на объектно-ориентированном языке программирования. Применение готовых реализаций алгоритмов защиты информации и библиотек, содержащих такие реализации, не допускается.

1. Реализуйте stateless-сервис, предоставляющий объектный функционал для:
    - сложения двоичных полиномов (далее - элементов) из *GF*(2<sup>8</sup>) ;
    - умножения элементов из *GF*(2<sup>8</sup>) по заданному модулю;
    - взятия обратного элемента для элемента из *GF*(2<sup>8</sup>) по заданному модулю;
    - проверки двоичного полинома степени 8 на неприводимость над *GF*(2<sup>8</sup>) ;
    - построения коллекции всех неприводимых над *GF*(2<sup>8</sup>) двоичных полиномов степени 8 (спойлер: их должно получиться 30);
    - построения разложения двоичного полинома произвольной степени на неприводимые множители из *GF*(2<sup>n</sup>), 𝑛 ∈ 𝑁
При попытке выполнения операции умножения/взятия обратного элемента по приводимому над *GF*(2<sup>8</sup>) модулю, генерируйте (и перехватывайте в вызывающем коде) исключительную ситуацию. Значения элементов из *GF*(2<sup>8</sup>) и модулей над *GF*(2<sup>8</sup>) передавайте и возвращайте в виде однобайтовых значений (byte, char, … (в зависимости от используемого языка программирования)). При вычислениях максимизируйте использование битовых операций.

2. На базе интерфейсов 2.1, 2.2, 2.3 (см. Задания к работе №1 по защите информации, задание 2) реализуйте класс, функционал которого позволяет выполнять [де]шифрование блока данных алгоритмом Rijndael. Обеспечьте возможность переиспользования для [де]шифрования различных блоков данных ключей раунда, полученных в результате выполнения процедуры расширения ключа. Реализация алгоритма должна поддерживать работу с блоками длиной 128/192/256 бит и ключами длиной 128/192/256 бит, а также предоставлять возможность настройки модуля над *GF*(2<sup>8</sup>) на этапе конструктора (используйте функционал, реализованный в задании 1). S-матрицы, необходимые для выполнения работы алгоритма, необходимо отложенно инициализировать для настроенного модуля над *GF*(2<sup>8</sup>). Вычисление прямой S-матрицы через обратную и наоборот не допускается. При работе с элементами из *GF*(2<sup>8</sup>) используйте функционал, реализованный в задании 1.

3. Продемонстрируйте выполнение шифрования и дешифрования псевдослучайных последовательностей байтов и файлов (текстовых, музыкальных, изображений, видео и т. д.) реализованным в задании 2 функционалом с использованием различных режимов шифрования и различных режимов набивки (см. Задания к работе №1 по защите информации, задание 4), различных длины блока и длины ключа, а также с использованием различных неприводимых над *GF*(2<sup>8</sup>) двоичных полиномов степени 8.

<a name="English"></a>
## English
[Перейти к русскому](#Russian)

(translated by AI)

All tasks are to be completed in an object-oriented programming language. The use of ready-made implementations of information protection algorithms and libraries containing such implementations is not allowed.

1. Implement a stateless service that provides object-oriented functionality for:
    - adding binary polynomials (elements) from *GF*(2<sup>8</sup>);
    - multiplying elements from *GF*(2<sup>8</sup>) by a given modulus;
    - taking the inverse element for an element from *GF*(2<sup>8</sup>) by a given modulus;
    - checking a binary polynomial of degree 8 for irreducibility over *GF*(2<sup>8</sup>);
    - constructing a collection of all irreducible binary polynomials of degree 8 over *GF*(2<sup>8</sup>) (spoiler: there should be 30 of them);
    - constructing the factorization of a binary polynomial of arbitrary degree into irreducible factors from *GF*(2<sup>n</sup>), 𝑛 ∈ 𝑁.
When attempting to perform a multiplication/inverse element operation by a reducible modulus over *GF*(2<sup>8</sup>), generate (and catch in the calling code) an exceptional situation. Values of elements from *GF*(2<sup>8</sup>) and moduli over *GF*(2<sup>8</sup>) should be passed and returned as single-byte values (byte, char, … (depending on the programming language used)). When performing calculations, maximize the use of bitwise operations.

2. Based on interfaces 2.1, 2.2, 2.3 (see Tasks for Work No. 1 on Information Protection, Task 2), implement a class whose functionality allows for [de]encryption of a block of data using the Rijndael algorithm. Ensure the ability to reuse round keys obtained as a result of the key expansion procedure for [de]encrypting different blocks of data. The implementation of the algorithm should support working with blocks of length 128/192/256 bits and keys of length 128/192/256 bits, as well as provide the ability to configure the modulus over *GF*(2<sup>8</sup>) during the constructor phase (use the functionality implemented in Task 1). The S-boxes required for the algorithm to work should be lazily initialized for the configured modulus over *GF*(2<sup>8</sup>). Calculating the direct S-box through the inverse and vice versa is not allowed. When working with elements from *GF*(2<sup>8</sup>), use the functionality implemented in Task 1.

3. Demonstrate the execution of encryption and decryption of pseudorandom byte sequences and files (text, music, images, video, etc.) using the functionality implemented in Task 2 with different encryption modes and padding modes (see Tasks for Work No. 1 on Information Protection, Task 4), different block lengths and key lengths, as well as using different irreducible binary polynomials of degree 8 over *GF*(2<sup>8</sup>).