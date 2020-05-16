---
title: Device Management Enhancement
authors:
    - "@luogangyi"
approvers:
creation-date: 2020-05-13
last-updated: 
status: 
---

# Device Management Enhancement
## Motivation

Device management is a key feature required for IoT use-cases in edge computing.
This proposal addresses how can we enhance device management including more flexible
protocol support, more reasonable device model and device instance design, and
additional way to handle data which is collected from device.

### Goals

* Add customized protocol support
* Add ability to allow user get and process data in edge node
* Improve device model and device instance crd design

### Non-goals

* To support streaming data from video device
* To provide specific protocols and mappers

## Proposal

We propose 4 modifications on current design.
* Move property visitors from device model to device instance.
* Add collectCycle and reportCycle under property visitor.
* Add data section besides twin section.
* Add customized protocol definition in device model and device instance CRDs.

### Use Cases

* Reuse device model. 
  * Considering device properties are physical attributes, but property visitors are manually configured
attributes. Combining device properties and property visitors in device model reduce the reusability of device model.
  * Case 1: Same devices are connected to a central management server, eg. SCADA. In this case, devices have same properties but
  different property visitors.
  * Case 2: Same devices are using different industrial protocol. In this case, devices have same properties but
  different property visitors.
* Customized data collect cycle and report cycle
   * Users can define collect cycle and report cycle to each property. For example, a temperature property may need be collected
   per second, while a throughput property may need be collected per hour.
* Deal data of non-twin properties.
  * Currently, only twin properties will be sync between edge and cloud. Non-twin properties are not processed by edge-core. 
  Time-Serial data are produced from devices and should have a way to allow user deal with these data.
* Deal various industrial protocols
  * Currently, only Modbus, OPC-UA and bluetooth are supported by KubeEdge. However there are thousands of industrial protocols.
  It is impossible to define all these protocols in KubeEdge. If users want to use these un-predefined protocols, we should provide
  a way to support.

## Design Details

### Move property visitors from device model to device instance.

- move property visitors of device CRD to device instance CRD
- move property visitors of DeviceModelSpec to DeviceSpec struct.
- change device profile generating procedure

### Add collectCycle and reportCycle under property visitor.
- add collectCycle and reportCycle under property visitor in device instance CRD.
- add collectCycle and reportCycle in DevicePropertyVisitor struct.

### Add data section besides twin section.
- add data section in device instance CRD.
- add DeviceData in Device struct.
- inject data section and twin section into configmap
- add new MQTT topic to handle data from data section

### Add customized protocols support
- add 'other' under property visitor, type is object

### New device model CRD sample
```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: devicemodels.devices.kubeedge.io
spec:
  group: devices.kubeedge.io
  names:
    kind: DeviceModel
    plural: devicemodels
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          properties:
            properties:
              description: 'Required: List of device properties.'
              items:
                properties:
                  description:
                    description: The device property description.
                    type: string
                  name:
                    description: 'Required: The device property name.'
                    type: string
                  type:
                    description: 'Required: PropertyType represents the type and data
                      validation of the property.'
                    properties:
                      int:
                        properties:
                          accessMode:
                            description: 'Required: Access mode of property, ReadWrite
                              or ReadOnly.'
                            type: string
                            enum:
                              - ReadOnly
                              - ReadWrite
                          defaultValue:
                            format: int64
                            type: integer
                          maximum:
                            format: int64
                            type: integer
                          minimum:
                            format: int64
                            type: integer
                          unit:
                            description: The unit of the property
                            type: string
                        required:
                          - accessMode
                        type: object
                      string:
                        properties:
                          accessMode:
                            description: 'Required: Access mode of property, ReadWrite
                              or ReadOnly.'
                            type: string
                            enum:
                              - ReadOnly
                              - ReadWrite
                          defaultValue:
                            type: string
                        required:
                          - accessMode
                        type: object
                    type: object
                required:
                  - name
                  - type
                type: object
              type: array
          type: object
  version: v1alpha1
```

### New device instance CRD sample

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: devices.devices.kubeedge.io
spec:
  group: devices.kubeedge.io
  names:
    kind: Device
    plural: devices
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          properties:
            deviceModelRef:
              description: 'Required: DeviceModelRef is reference to the device model
                used as a template to create the device instance.'
              type: object
            nodeSelector:
              description: NodeSelector indicates the binding preferences between
                devices and nodes. Refer to k8s.io/kubernetes/pkg/apis/core NodeSelector
                for more details
              type: object
            protocol:
              description: 'Required: The protocol configuration used to connect to
                the device.'
              properties:
                bluetooth:
                  description: Protocol configuration for bluetooth
                  properties:
                    macAddress:
                      description: Unique identifier assigned to the device.
                      type: string
                  type: object
                modbus:
                  description: Protocol configuration for modbus
                  properties:
                    rtu:
                      properties:
                        baudRate:
                          description: Required. BaudRate 115200|57600|38400|19200|9600|4800|2400|1800|1200|600|300|200|150|134|110|75|50
                          format: int64
                          type: integer
                          enum:
                            - 115200
                            - 57600
                            - 38400
                            - 19200
                            - 9600
                            - 4800
                            - 2400
                            - 1800
                            - 1200
                            - 600
                            - 300
                            - 200
                            - 150
                            - 134
                            - 110
                            - 75
                            - 50
                        dataBits:
                          description: Required. Valid values are 8, 7, 6, 5.
                          format: int64
                          type: integer
                          enum:
                            - 8
                            - 7
                            - 6
                            - 5
                        parity:
                          description: Required. Valid options are "none", "even",
                            "odd". Defaults to "none".
                          type: string
                          enum:
                            - none
                            - even
                            - odd
                        serialPort:
                          description: Required.
                          type: string
                        slaveID:
                          description: Required. 0-255
                          format: int64
                          type: integer
                          minimum: 0
                          maximum: 255
                        stopBits:
                          description: Required. Bit that stops 1|2
                          format: int64
                          type: integer
                          enum:
                            - 1
                            - 2
                      required:
                        - baudRate
                        - dataBits
                        - parity
                        - serialPort
                        - slaveID
                        - stopBits
                      type: object
                    tcp:
                      properties:
                        ip:
                          description: Required.
                          type: string
                        port:
                          description: Required.
                          format: int64
                          type: integer
                        slaveID:
                          description: Required.
                          type: string
                      required:
                        - ip
                        - port
                        - slaveID
                      type: object
                  type: object
                opcua:
                  description: Protocol configuration for opc-ua
                  properties:
                    certificate:
                      description: Certificate for access opc server.
                      type: string
                    password:
                      description: Password for access opc server.
                      type: string
                    privateKey:
                      description: PrivateKey for access opc server.
                      type: string
                    securityMode:
                      description: Defaults to "none".
                      type: string
                    securityPolicy:
                      description: Defaults to "none".
                      type: string
                    timeout:
                      description: Timeout seconds for the opc server connection.???
                      format: int64
                      type: integer
                    url:
                      description: 'Required: The URL for opc server endpoint.'
                      type: string
                    userName:
                      description: Username for access opc server.
                      type: string
                  required:
                    - url
                  type: object
              type: object
            propertyVisitors:
              description: 'Required: List of property visitors which describe how
                to access the device properties. PropertyVisitors must unique by propertyVisitor.propertyName.'
              items:
                properties:
                  bluetooth:
                    description: Bluetooth represents a set of additional visitor
                      config fields of bluetooth protocol.
                    properties:
                      characteristicUUID:
                        description: 'Required: Unique ID of the corresponding operation'
                        type: string
                      dataConverter:
                        description: Responsible for converting the data being read
                          from the bluetooth device into a form that is understandable
                          by the platform
                        properties:
                          endIndex:
                            description: 'Required: Specifies the end index of incoming
                              byte stream to be considered to convert the data the
                              value specified should be inclusive for example if 3
                              is specified it includes the third index'
                            format: int64
                            type: integer
                          orderOfOperations:
                            description: Specifies in what order the operations(which
                              are required to be performed to convert incoming data
                              into understandable form) are performed
                            items:
                              properties:
                                operationType:
                                  description: 'Required: Specifies the operation
                                    to be performed to convert incoming data'
                                  type: string
                                  enum:
                                    - Add
                                    - Subtract
                                    - Multiply
                                    - Divide
                                operationValue:
                                  description: 'Required: Specifies with what value
                                    the operation is to be performed'
                                  format: double
                                  type: number
                              type: object
                            type: array
                          shiftLeft:
                            description: Refers to the number of bits to shift left,
                              if left-shift operation is necessary for conversion
                            format: int64
                            type: integer
                          shiftRight:
                            description: Refers to the number of bits to shift right,
                              if right-shift operation is necessary for conversion
                            format: int64
                            type: integer
                          startIndex:
                            description: 'Required: Specifies the start index of the
                              incoming byte stream to be considered to convert the
                              data. For example: start-index:2, end-index:3 concatenates
                              the value present at second and third index of the incoming
                              byte stream. If we want to reverse the order we can
                              give it as start-index:3, end-index:2'
                            format: int64
                            type: integer
                        required:
                          - endIndex
                          - startIndex
                        type: object
                      dataWrite:
                        description: 'Responsible for converting the data coming from
                          the platform into a form that is understood by the bluetooth
                          device For example: "ON":[1], "OFF":[0]'
                        type: object
                    required:
                      - characteristicUUID
                    type: object
                  modbus:
                    description: Modbus represents a set of additional visitor config
                      fields of modbus protocol.
                    properties:
                      isRegisterSwap:
                        description: Indicates whether the high and low register swapped.
                          Defaults to false.
                        type: boolean
                      isSwap:
                        description: Indicates whether the high and low byte swapped.
                          Defaults to false.
                        type: boolean
                      limit:
                        description: 'Required: Limit number of registers to read/write.'
                        format: int64
                        type: integer
                      offset:
                        description: 'Required: Offset indicates the starting register
                          number to read/write data.'
                        format: int64
                        type: integer
                      register:
                        description: 'Required: Type of register'
                        type: string
                        enum:
                          - CoilRegister
                          - DiscreteInputRegister
                          - InputRegister
                          - HoldingRegister
                      scale:
                        description: The scale to convert raw property data into final
                          units. Defaults to 1.0
                        format: double
                        type: number
                    required:
                      - limit
                      - offset
                      - register
                    type: object
                  opcua:
                    description: Opcua represents a set of additional visitor config
                      fields of opc-ua protocol.
                    properties:
                      browseName:
                        description: The name of opc-ua node
                        type: string
                      nodeID:
                        description: 'Required: The ID of opc-ua node, e.g. "ns=1,i=1005"'
                        type: string
                    required:
                      - nodeID
                    type: object
                  customizedProtocol:
                    description: customized protocol
                    properties:
                      protocalName:
                        description: The name of protocol
                        type: string
                      definition:
                        description: customized definition
                        type: object
                    required:
                      - protocolName
                      - definition
                    type: object
                  propertyName:
                    description: 'Required: The device property name to be accessed.
                      This should refer to one of the device properties defined in
                      the device model.'
                    type: string
                required:
                  - propertyName
                type: object
              type: array
          required:
            - deviceModelRef
            - propertyVisitors
          type: object
        status:
          properties:
            twins:
              description: A list of device twins containing desired/reported desired/reported
                values of twin properties. A passive device won't have twin properties
                and this list could be empty.
              items:
                properties:
                  desired:
                    description: 'Required: the desired property value.'
                    properties:
                      metadata:
                        description: Additional metadata like timestamp when the value
                          was reported etc.
                        type: object
                      value:
                        description: 'Required: The value for this property.'
                        type: string
                    required:
                      - value
                    type: object
                  propertyName:
                    description: 'Required: The property name for which the desired/reported
                      values are specified. This property should be present in the
                      device model.'
                    type: string
                  reported:
                    description: 'Required: the reported property value.'
                    properties:
                      metadata:
                        description: Additional metadata like timestamp when the value
                          was reported etc.
                        type: object
                      value:
                        description: 'Required: The value for this property.'
                        type: string
                    required:
                      - value
                    type: object
                required:
                  - propertyName
                type: object
              type: array
            data:
              description: A list of device properties which contain time-serial data
              items:
                properties:
                  propertyName:
                    type: string
                  metadata:
                    description: Any additional params
                    type: string
                required:
                  - propertyName
                type: object
              type: array
          type: object
  version: v1alpha1
```
### New configMap sample

To avoid duplicated property name and protocol configuration, we move property visitor section and protocol configuration section into device instance section.

```yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: device-profile-config-01 // needs to be generated by device controller.
  namespace: foo
data:
  deviceProfile.json: |-
{
  "deviceInstances": [{
    "id": "1",
    "name": "device1",
    "model": "SensorTagModel",
    "protocol": {
      "name": "modbus-rtu-01",
      "protocol": "modbus-rtu",
      "protocolConfig": {
        "serialPort": "1",
        "baudRate": "115200",
        "dataBits": "8",
        "parity": "even",
        "stopBits": "1",
        "slaveID": "1"
      }
    },
    "propertyVisitors": [{
        "name": "temperature",
        "propertyName": "temperature",
        "modelName": "SensorTagModel",
        "protocol": "modbus-rtu",
        "visitorConfig": {
          "register": "CoilRegister",
          "offset": "2",
          "limit": "1",
          "scale": "1.0",
          "isSwap": "true",
          "isRegisterSwap": "true"
        }
      },
      {
        "name": "temperatureEnable",
        "propertyName": "temperature-enable",
        "modelName": "SensorTagModel",
        "protocol": "modbus-rtu",
        "visitorConfig": {
          "register": "DiscreteInputRegister",
          "offset": "3",
          "limit": "1",
          "scale": "1.0",
          "isSwap": "true",
          "isRegisterSwap": "true"
        }
      }
    ]
  }],
  "deviceModels": [{
    "name": "SensorTagModel",
    "description": "TI Simplelink SensorTag Device Attributes Model",
    "properties": [{
        "name": "temperature",
        "datatype": "int",
        "accessMode": "r",
        "unit": "Degree Celsius",
        "maximum": "100"
      },
      {
        "name": "temperature-enable",
        "datatype": "string",
        "accessMode": "rw",
        "defaultValue": "OFF"
      }
    ]
  }]
}
```

## Open questions
- Should we split the monolithic configmap, let each mapper has its own configmap?
- Should we add a default collect cycle and report cycle in device instance?