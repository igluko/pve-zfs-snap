# pve-zfs-snap
## Описание
Эта программа создает и ротирует ZFS снимки, но только для запущенных VM и контейнеров.
Все операции с ZFS выполняются через zfs program, с использованием lua скриптов.
Это позволяет делать множество операций в рамках 1 ZFS транзакции (атомарно).
Например, все снимки будут выполнены разом и иметь один номер транзакции.

Программа предназначена для запуска из крона.

Причина создания данной программы - поддержка 3 сторонней репликации VM:
- B <- A -> C (рабочее состояние)
- B -> A -> C (фаэйловер, переключение на B)

Особенности репликации на 3ю сторону (-> C) :
- нельзя создавать служебные снимки, нужно реплицировать имеющиеся снимки
- нужно создавать закладки
- переодически снимки должны создаваться только для запущенных VM

## Параметры запуска
Минимальное число параметров - 1

Все параметры параметры должены иметь формат `<one_letter><int>`

Пример использования: `./pve-zfs-snap f100000`

Возожные ключи и их описания:
- `f<int>` - количество frequently снимков (создаются при каждом запуске)
- `h<int>` - количество hourly снимков
- `d<int>` - количество daily снимков
- `m<int>` - количество monthly снимков
- `y<int>` - количество yearly снимков

## Cron
Программа автоматически регистрирует свой исполняемый файл в cron с теми параметрами, которые были переданы во время запуска.
Частота запуска 1 раз в 15 минут. 
Регистрация в cron происходит только во время интерактивного запуска, при запуске из cron, этот этап пропускается.

## Отслеживание start/stop VM
На резервной площадке мы проверяем дату последнего снимка каждого диска. Если снимок был сделан давно, мы алертим.

Пока снимками занимался sanoid это работало отлично, т.к. он делает снимки всегда,
независимо от того запущена ли VM, но теперь мы делаем снимки только когда VM запущена.

Это значит, что когда VM остановлена, на резервную площадку перестанут приходить новые снимки и система мониторинга начнет ругаться.

Чтобы этого избежать, мы должны делать 1 принудительный снимок c суффиксом `stopped` после остановки VM. 
Система мониторинга не должна алертить, если последний снимок содержит суффикс stopped

Мы должны отслеживать остановки и запуски VM, чтобы делать 1 принудительный снимков после каждой остановки VM
- Это значит, что нам нужно отслеживать изменение состояния, а для этого нужно где то хранить старое состояние, причем так, чтобы оно было доступно и на репликах
- Мы должны сохранять hostname, на которой была запущена VM, т.к. поменять это поле на stopped может только на той же машине

Мы будем использовать пользовательское поле label:running внутри ZFS, чтобы сохранять и обновлять состояние
* Если VM запущена, установим поле `label:running=${hostname}`
* Если VM остановлена, а поле `label:running` равно `-` или `${hostname}`, то поменяем значение поля на `stopped`
В противном случае менять не будем, т.к. там скорее всего чужой hostname, то есть VM запущена на другом сервере.
* Мы должны менять значение поля, только если оно в этом нуждается

### stopped > running
- Найдем все датасеты, которые должны приобрести свойство `label:running=${hostname}`
  - загружаем все датасеты
	- фильтруем: оставляем только датасеты VMID running
	- фильтруем: оставляем только датасеты без свойства `label:running=${hostname}`
- Обнуляем свойство `label:running` у отфильтрованных датасетов
	- `zfs set label:running=${hostname} mypool/mydataset`

### running -> stopped
- Найдем все датасеты, которые должны приобрести свойство `label:running=stopped`
	- загружаем все датасеты
	- фильтруем: оставляем только датасеты связанные с VM или контейнерами
	- фильтруем: оставляем только датасеты VMID которых не running
	- фильтруем: оставляем только датасеты со свойством `label:running=${hostname}` или `label:running="-"`
- Обнуляем свойство label:running у отфильтрованных датасетов
	- `zfs set label:running=stopped mypool/mydataset`
- Делаем снимок этих датасетов
	- `zfs snapshot dataset@autosnap_${TIME}_stopped`
		- Диски остановленных VM уже должны находиться в консистентном состоянии