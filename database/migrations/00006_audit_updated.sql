-- +goose Up
	
ALTER TABLE FS_ContainerRelate ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_FieldScoutingLog ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_HabitatRelate ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_InspectionSample ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_InspectionSampleDetail ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_LineLocation ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_LocationTracking ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_MosquitoInspection ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_PointLocation ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_PolygonLocation ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_Pool ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_PoolDetail ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_ProposedTreatmentArea ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_QAMosquitoInspection ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_RodentLocation ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_SampleCollection ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_SampleLocation ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_ServiceRequest ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_SpeciesAbundance ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_StormDrain ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_TimeCard ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_TrapData ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_TrapLocation ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_Treatment ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_TreatmentArea ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_Zones ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;
ALTER TABLE FS_Zones2 ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT current_timestamp;

-- +goose Down

ALTER TABLE FS_ContainerRelate DROP COLUMN updated;
ALTER TABLE FS_FieldScoutingLog DROP COLUMN updated;
ALTER TABLE FS_HabitatRelate DROP COLUMN updated;
ALTER TABLE FS_InspectionSample DROP COLUMN updated;
ALTER TABLE FS_InspectionSampleDetail DROP COLUMN updated;
ALTER TABLE FS_LineLocation DROP COLUMN updated;
ALTER TABLE FS_LocationTracking DROP COLUMN updated;
ALTER TABLE FS_MosquitoInspection DROP COLUMN updated;
ALTER TABLE FS_PointLocation DROP COLUMN updated;
ALTER TABLE FS_PolygonLocation DROP COLUMN updated;
ALTER TABLE FS_Pool DROP COLUMN updated;
ALTER TABLE FS_PoolDetail DROP COLUMN updated;
ALTER TABLE FS_ProposedTreatmentArea DROP COLUMN updated;
ALTER TABLE FS_QAMosquitoInspection DROP COLUMN updated;
ALTER TABLE FS_RodentLocation DROP COLUMN updated;
ALTER TABLE FS_SampleCollection DROP COLUMN updated;
ALTER TABLE FS_SampleLocation DROP COLUMN updated;
ALTER TABLE FS_ServiceRequest DROP COLUMN updated;
ALTER TABLE FS_SpeciesAbundance DROP COLUMN updated;
ALTER TABLE FS_StormDrain DROP COLUMN updated;
ALTER TABLE FS_TimeCard DROP COLUMN updated;
ALTER TABLE FS_TrapData DROP COLUMN updated;
ALTER TABLE FS_TrapLocation DROP COLUMN updated;
ALTER TABLE FS_Treatment DROP COLUMN updated;
ALTER TABLE FS_TreatmentArea DROP COLUMN updated;
ALTER TABLE FS_Zones DROP COLUMN updated;
ALTER TABLE FS_Zones2 DROP COLUMN updated;

